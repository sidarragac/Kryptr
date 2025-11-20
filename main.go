//go:build linux
// +build linux

package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"unsafe"
)

// ----------------------------------------------------------------------
// FLAGS Y MAIN
// ----------------------------------------------------------------------

func main() {
	cFlag := flag.Bool("c", false, "Comprimir archivo")
	dFlag := flag.Bool("d", false, "Descomprimir archivo")
	eFlag := flag.Bool("e", false, "Encriptar archivo")
	uFlag := flag.Bool("u", false, "Desencriptar archivo")
	compFlag := flag.String("comp-alg", "", "Nombre del algoritmo de compresión (huff)")
	encFlag := flag.String("enc-alg", "", "Nombre del algoritmo de encriptación (xor)")
	iFlag := flag.String("i", "", "Ruta del archivo o directorio de entrada")
	oFlag := flag.String("o", "", "Ruta del archivo o directorio de salida")

	flag.Parse()

	if *iFlag == "" {
		fmt.Println("Debes especificar la ruta de entrada con -i")
		return
	}

	// Determinar si la ruta es archivo o directorio
	var st syscall.Stat_t
	err := syscall.Stat(*iFlag, &st)
	if err != nil {
		fmt.Printf("Error al acceder a %s: %v\n", *iFlag, err)
		return
	}

	var wg sync.WaitGroup
	sem := make(chan struct{}, 16) // limitar goroutines simultáneas

	// Si es directorio → recorrer recursivamente
	if st.Mode&syscall.S_IFMT == syscall.S_IFDIR {
		fd, err := syscall.Open(*iFlag, syscall.O_RDONLY|syscall.O_DIRECTORY, 0)
		if err != nil {
			fmt.Printf("No se pudo abrir directorio %s: %v\n", *iFlag, err)
			return
		}
		defer syscall.Close(fd)

		wg.Add(1)
		go func() {
			defer wg.Done()
			recorrerDir(*iFlag, fd, *cFlag, *dFlag, *eFlag, *uFlag, *compFlag, *encFlag, *oFlag, &wg, sem)
		}()

		wg.Wait()
		fmt.Println("Procesamiento completo.")
		return
	}

	// Si es archivo → procesar directamente
	procesarArchivo(*iFlag, *oFlag, *cFlag, *dFlag, *eFlag, *uFlag, *compFlag, *encFlag)
}

// ----------------------------------------------------------------------
// FUNCIONES DE PROCESAMIENTO
// ----------------------------------------------------------------------

func procesarArchivo(path string, out string, c, d, e, u bool, compAlg, encAlg string) {
	if c || compAlg == "huff" {
		comprimir(path, out)
	}
	if d {
		descomprimir(path, out)
	}
	if e || encAlg == "xor" {
		Encriptar(path, out)
	}
	if u {
		Desencriptar(path, out)
	}
}

func comprimir(file string, out string) {
	fmt.Println("Comprimiendo " + file)
	data, err := ioutil.ReadFile(file)
	if err != nil {
		fmt.Printf("Error leyendo %s: %v\n", file, err)
		return
	}

	fmt.Printf("Tamaño original: %d bytes\n", len(data))

	// empaquetar incluyendo nombre original
	compressed := PackWithMeta(data, filepath.Base(file))
	fmt.Printf("Tamaño comprimido: %d bytes\n", len(compressed))

	// determinar ruta de salida (.bin por defecto)
	outPath := out
	if outPath == "" {
		// reemplazar extensión por .bin
		ext := filepath.Ext(file)
		if ext == "" {
			outPath = file + ".bin"
		} else {
			outPath = strings.TrimSuffix(file, ext) + ".bin"
		}
	} else {
		// si out es un directorio, usar basename sin extensión + .bin
		fi, err := os.Stat(outPath)
		if err == nil && fi.IsDir() {
			base := filepath.Base(file)
			ext := filepath.Ext(base)
			name := strings.TrimSuffix(base, ext)
			outPath = filepath.Join(outPath, name+".bin")
		}
	}

	if err := ioutil.WriteFile(outPath, compressed, 0644); err != nil {
		fmt.Printf("Error escribiendo %s: %v\n", outPath, err)
		return
	}
	fmt.Printf("Guardado: %s\n", outPath)
}

func descomprimir(file string, out string) {
	fmt.Println("Descomprimiendo " + file)
	data, err := ioutil.ReadFile(file)
	if err != nil {
		fmt.Printf("Error leyendo %s: %v\n", file, err)
		return
	}

	// intentar extraer metadata
	origName, payload := UnpackWithMeta(data)
	decompressed := huffmanDecompress(payload)
	if decompressed == nil {
		fmt.Printf("Error: no se pudo descomprimir %s\n", file)
		return
	}
	fmt.Printf("Tamaño comprimido (entrada): %d bytes\n", len(data))
	fmt.Printf("Tamaño descomprimido: %d bytes\n", len(decompressed))
	outPath := out
	if outPath == "" {
		if origName != "" {
			outPath = filepath.Join(filepath.Dir(file), origName)
		} else {
			// intentar restaurar nombre original quitando sufijo conocido
			suffixes := []string{".bin", ".kry", ".huff"}
			restored := ""
			for _, s := range suffixes {
				if strings.HasSuffix(file, s) {
					restored = strings.TrimSuffix(file, s)
					break
				}
			}
			if restored != "" {
				outPath = restored
			} else {
				outPath = file + ".dec"
			}
		}
	} else {
		fi, err := os.Stat(outPath)
		if err == nil && fi.IsDir() {
			// si es directorio, preferir el nombre original si está en la metadata
			if origName != "" {
				outPath = filepath.Join(outPath, origName)
			} else {
				// intentar usar basename sin sufijo
				base := filepath.Base(file)
				for _, s := range []string{".bin", ".kry", ".huff"} {
					if strings.HasSuffix(base, s) {
						base = strings.TrimSuffix(base, s)
						break
					}
				}
				outPath = filepath.Join(outPath, base)
			}
		}
	}

	if err := ioutil.WriteFile(outPath, decompressed, 0644); err != nil {
		fmt.Printf("Error escribiendo %s: %v\n", outPath, err)
		return
	}
	fmt.Printf("Guardado: %s\n", outPath)
}

/*
func encriptar(file string) {
	fmt.Println("Encriptando " + file)
}

func desencriptar(file string) {
	fmt.Println("Desencriptando " + file)
}
*/

// ----------------------------------------------------------------------
// FUNCIONES DE EXPLORACIÓN CON SYSCALL
// ----------------------------------------------------------------------

func recorrerDir(path string, fd int, c, d, e, u bool, compAlg, encAlg, out string, wg *sync.WaitGroup, sem chan struct{}) {
	buf := make([]byte, 4096)
	for {
		n, _, errno := syscall.Syscall(syscall.SYS_GETDENTS64, uintptr(fd), uintptr(unsafe.Pointer(&buf[0])), uintptr(len(buf)))
		if errno != 0 {
			fmt.Printf("Error leyendo %s: %v\n", path, errno)
			return
		}
		if n == 0 {
			return // sin más entradas
		}

		b := buf[:n]
		for len(b) > 0 {
			dirent := (*linuxDirent64)(unsafe.Pointer(&b[0]))
			nameBytes := (*[256]byte)(unsafe.Pointer(&dirent.Name[0]))
			name := cstring(nameBytes[:])

			if name == "." || name == ".." {
				// saltar
			} else {
				fullPath := path + "/" + name
				if dirent.Type == syscall.DT_REG {
					// archivo regular
					wg.Add(1)
					sem <- struct{}{}
					go func(p, out, compAlg, encAlg string) {
						defer wg.Done()
						defer func() { <-sem }()
						procesarArchivo(p, out, c, d, e, u, compAlg, encAlg)
					}(fullPath, out, compAlg, encAlg)
				} else if dirent.Type == syscall.DT_DIR {
					// subdirectorio → recursión
					subFd, err := syscall.Open(fullPath, syscall.O_RDONLY|syscall.O_DIRECTORY, 0)
					if err == nil {
						wg.Add(1)
						go func(p string, f int, out, compAlg, encAlg string) {
							defer wg.Done()
							recorrerDir(p, f, c, d, e, u, compAlg, encAlg, out, wg, sem)
							syscall.Close(f)
						}(fullPath, subFd, out, compAlg, encAlg)
					}
				}
			}

			if dirent.Reclen == 0 || int(dirent.Reclen) > len(b) {
				break
			}
			b = b[dirent.Reclen:]
		}
	}
}

// Estructura de linux_dirent64 para getdents64
type linuxDirent64 struct {
	Ino    uint64
	Off    int64
	Reclen uint16
	Type   byte
	Name   [256]byte
}

// Convierte nombre C-string a Go string
func cstring(b []byte) string {
	for i, c := range b {
		if c == 0 {
			return string(b[:i])
		}
	}
	return string(b)
}
