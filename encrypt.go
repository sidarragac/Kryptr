package main

import (
	"fmt"
	"strings"
	"syscall"
)

// generateRoundKey creates a slightly modified version of the key for each round.
func generateRoundKey(baseKey []byte, round int) []byte {
	roundKey := make([]byte, len(baseKey))
	for i := range baseKey {
		roundKey[i] = baseKey[i] + byte(round)
	}
	return roundKey
}

func xorEncrypt(plaintext, key []byte, rounds int) []byte {
	data := make([]byte, len(plaintext))
	copy(data, plaintext)

	for r := 0; r < rounds; r++ {
		roundKey := generateRoundKey(key, r)
		for i := 0; i < len(data); i++ {
			data[i] ^= roundKey[i%len(roundKey)]
		}
	}
	return data
}

func xorDecrypt(ciphertext, key []byte, rounds int) []byte {
	data := make([]byte, len(ciphertext))
	copy(data, ciphertext)

	for r := rounds - 1; r >= 0; r-- {
		roundKey := generateRoundKey(key, r)
		for i := 0; i < len(data); i++ {
			data[i] ^= roundKey[i%len(roundKey)]
		}
	}
	return data
}

// baseName devuelve el basename de una ruta (sin usar path/filepath)
func baseName(p string) string {
	if p == "" {
		return ""
	}
	// remover trailing '/'
	for len(p) > 1 && p[len(p)-1] == '/' {
		p = p[:len(p)-1]
	}
	i := strings.LastIndexByte(p, '/')
	if i == -1 {
		return p
	}
	return p[i+1:]
}

// dirName devuelve el directorio padre (similar a filepath.Dir)
func dirName(p string) string {
	if p == "" {
		return "."
	}
	// remover trailing '/' salvo si es root
	for len(p) > 1 && p[len(p)-1] == '/' {
		p = p[:len(p)-1]
	}
	i := strings.LastIndexByte(p, '/')
	if i == -1 {
		return "."
	}
	if i == 0 {
		return "/"
	}
	return p[:i]
}

// existsDir comprueba si path existe y es directorio usando syscall
func existsDir(path string) (bool, error) {
	var st syscall.Stat_t
	if err := syscall.Stat(path, &st); err != nil {
		if err == syscall.ENOENT {
			return false, nil
		}
		return false, err
	}
	return (st.Mode & syscall.S_IFMT) == syscall.S_IFDIR, nil
}

// mkdirAll crea directorios recursivamente usando syscall.Mkdir
func mkdirAll(path string, mode uint32) error {
	if path == "" || path == "." {
		return nil
	}
	p := path
	// normalizar: remover trailing '/' excepto si es "/"
	for len(p) > 1 && p[len(p)-1] == '/' {
		p = p[:len(p)-1]
	}
	// construir progresivamente
	parts := strings.Split(p, "/")
	cur := ""
	if strings.HasPrefix(p, "/") {
		cur = "/"
	}
	for _, part := range parts {
		if part == "" {
			continue
		}
		if cur == "/" {
			cur = "/" + part
		} else if cur == "" {
			cur = part
		} else {
			cur = cur + "/" + part
		}
		// comprobar existencia
		dirExists, err := existsDir(cur)
		if err != nil {
			return err
		}
		if dirExists {
			continue
		}
		// intentar crear
		if err := syscall.Mkdir(cur, mode); err != nil {
			// EEXIST puede ocurrir por race; comprobar si ahora es directorio
			if err == syscall.EEXIST {
				dirExists2, err2 := existsDir(cur)
				if err2 != nil {
					return err2
				}
				if dirExists2 {
					continue
				}
			}
			return err
		}
	}
	return nil
}

func Encriptar(inPath, outPath string) {
	key := []byte("KEY")
	rounds := 5

	inPathNoExtension := strings.Split(inPath, ".")[0]

	// calcular outPath usando solo syscalls/helpers
	if outPath == "" {
		outPath = inPathNoExtension + ".kry"
	} else {
		// si outPath es directorio existente -> usar basename(inPathNoExtension)+".kry"
		if ok, err := existsDir(outPath); err == nil && ok {
			outPath = strings.TrimRight(outPath, "/") + "/" + baseName(inPathNoExtension) + ".kry"
		} else if len(outPath) > 0 && outPath[len(outPath)-1] == '/' {
			// termina en '/' -> crear directorio y usar basename
			if err := mkdirAll(outPath, 0755); err != nil {
				fmt.Printf("No se pudo crear directorio %s: %v\n", outPath, err)
				return
			}
			outPath = strings.TrimRight(outPath, "/") + "/" + baseName(inPathNoExtension) + ".kry"
		} else {
			// outPath es un archivo: asegurar que su dir padre exista
			dir := dirName(outPath)
			if dir != "." {
				if err := mkdirAll(dir, 0755); err != nil {
					fmt.Printf("No se pudo crear directorio %s: %v\n", dir, err)
					return
				}
			}
		}
	}

	// Abrir archivo de entrada (syscall)
	fd, err := syscall.Open(inPath, syscall.O_RDONLY, 0)
	if err != nil {
		fmt.Printf("Error abriendo %s: %v\n", inPath, err)
		return
	}
	defer syscall.Close(fd)

	// Obtener tamaño con fstat
	var st syscall.Stat_t
	if err := syscall.Fstat(fd, &st); err != nil {
		fmt.Printf("Fstat falló para %s: %v\n", inPath, err)
		return
	}
	size := int(st.Size)
	data := make([]byte, size)

	// Leer todo el archivo usando syscall.Read
	off := 0
	for off < size {
		n, err := syscall.Read(fd, data[off:])
		if n > 0 {
			off += n
		}
		if err != nil {
			if err == syscall.EINTR {
				continue
			}
			fmt.Printf("Error leyendo %s: %v\n", inPath, err)
			return
		}
		if n == 0 {
			break
		}
	}

	ciphertext := xorEncrypt(data, key, rounds)

	// Asegurar que el directorio padre del outPath exista (por si cambió)
	parent := dirName(outPath)
	if parent != "." {
		if err := mkdirAll(parent, 0755); err != nil {
			fmt.Printf("No se pudo crear directorio %s: %v\n", parent, err)
		}
	}

	// Escribir archivo de salida con syscall.Open/Write
	outFd, err := syscall.Open(outPath, syscall.O_WRONLY|syscall.O_CREAT|syscall.O_TRUNC, 0644)
	if err != nil {
		fmt.Printf("Error creando %s: %v\n", outPath, err)
		return
	}
	defer syscall.Close(outFd)

	written := 0
	for written < len(ciphertext) {
		n, err := syscall.Write(outFd, ciphertext[written:])
		if n > 0 {
			written += n
		}
		if err != nil {
			if err == syscall.EINTR {
				continue
			}
			fmt.Printf("Error escribiendo %s: %v\n", outPath, err)
			return
		}
	}
	fmt.Println("Encriptado ->", outPath)
}

func Desencriptar(inPath, outPath string) {
	key := []byte("KEY")
	rounds := 5

	inPathNoExtension := strings.Split(inPath, ".")[0]

	if outPath == "" {
		outPath = inPathNoExtension + ".dec"
	} else {
		if ok, err := existsDir(outPath); err == nil && ok {
			outPath = strings.TrimRight(outPath, "/") + "/" + baseName(inPathNoExtension) + ".dec"
		} else if len(outPath) > 0 && outPath[len(outPath)-1] == '/' {
			if err := mkdirAll(outPath, 0755); err != nil {
				fmt.Printf("No se pudo crear directorio %s: %v\n", outPath, err)
				return
			}
			outPath = strings.TrimRight(outPath, "/") + "/" + baseName(inPathNoExtension) + ".dec"
		} else {
			dir := dirName(outPath)
			if dir != "." {
				if err := mkdirAll(dir, 0755); err != nil {
					fmt.Printf("No se pudo crear directorio %s: %v\n", dir, err)
					return
				}
			}
		}
	}

	// Abrir archivo cifrado
	fd, err := syscall.Open(inPath, syscall.O_RDONLY, 0)
	if err != nil {
		fmt.Printf("Error abriendo %s: %v\n", inPath, err)
		return
	}
	defer syscall.Close(fd)

	var st syscall.Stat_t
	if err := syscall.Fstat(fd, &st); err != nil {
		fmt.Printf("Fstat falló para %s: %v\n", inPath, err)
		return
	}
	size := int(st.Size)
	data := make([]byte, size)

	off := 0
	for off < size {
		n, err := syscall.Read(fd, data[off:])
		if n > 0 {
			off += n
		}
		if err != nil {
			if err == syscall.EINTR {
				continue
			}
			fmt.Printf("Error leyendo %s: %v\n", inPath, err)
			return
		}
		if n == 0 {
			break
		}
	}

	plaintext := xorDecrypt(data, key, rounds)

	parent := dirName(outPath)
	if parent != "." {
		if err := mkdirAll(parent, 0755); err != nil {
			fmt.Printf("No se pudo crear directorio %s: %v\n", parent, err)
		}
	}

	outFd, err := syscall.Open(outPath, syscall.O_WRONLY|syscall.O_CREAT|syscall.O_TRUNC, 0644)
	if err != nil {
		fmt.Printf("Error creando %s: %v\n", outPath, err)
		return
	}
	defer syscall.Close(outFd)

	written := 0
	for written < len(plaintext) {
		n, err := syscall.Write(outFd, plaintext[written:])
		if n > 0 {
			written += n
		}
		if err != nil {
			if err == syscall.EINTR {
				continue
			}
			fmt.Printf("Error escribiendo %s: %v\n", outPath, err)
			return
		}
	}
	fmt.Println("Desencriptado ->", outPath)
}
