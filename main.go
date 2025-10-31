//go:build linux
// +build linux

package main

import (
	"flag"
	"fmt"
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
	iFlag := flag.String("i", "", "Ruta del archivo o directorio de entrada")

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
			recorrerDir(*iFlag, fd, *cFlag, *dFlag, *eFlag, *uFlag, &wg, sem)
		}()

		wg.Wait()
		fmt.Println("Procesamiento completo.")
		return
	}

	// Si es archivo → procesar directamente
	procesarArchivo(*iFlag, *cFlag, *dFlag, *eFlag, *uFlag)
}

// ----------------------------------------------------------------------
// FUNCIONES DE PROCESAMIENTO
// ----------------------------------------------------------------------

func procesarArchivo(path string, c, d, e, u bool) {
	if c {
		comprimir(path)
	}
	if d {
		descomprimir(path)
	}
	if e {
		encriptar(path)
	}
	if u {
		desencriptar(path)
	}
}

func comprimir(file string) {
	fmt.Println("Comprimiendo " + file)
}

func descomprimir(file string) {
	fmt.Println("Descomprimiendo " + file)
}

func encriptar(file string) {
	fmt.Println("Encriptando " + file)
}

func desencriptar(file string) {
	fmt.Println("Desencriptando " + file)
}

// ----------------------------------------------------------------------
// FUNCIONES DE EXPLORACIÓN CON SYSCALL
// ----------------------------------------------------------------------

func recorrerDir(path string, fd int, c, d, e, u bool, wg *sync.WaitGroup, sem chan struct{}) {
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
					go func(p string) {
						defer wg.Done()
						defer func() { <-sem }()
						procesarArchivo(p, c, d, e, u)
					}(fullPath)
				} else if dirent.Type == syscall.DT_DIR {
					// subdirectorio → recursión
					subFd, err := syscall.Open(fullPath, syscall.O_RDONLY|syscall.O_DIRECTORY, 0)
					if err == nil {
						wg.Add(1)
						go func(p string, f int) {
							defer wg.Done()
							recorrerDir(p, f, c, d, e, u, wg, sem)
							syscall.Close(f)
						}(fullPath, subFd)
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
