# Kryptr
Proyecto Final correspondiente a la materia Sistemas Operativos (SI2004) de la Universidad EAFIT


Desarrollado por:
- Mateo Pineda Álvarez
- Esteban Álvarez Zuluaga
- Santiago Idárraga Ceballos

## Requisitos técnicos:
- Sistema Operativo Linux
- GoLang Versión 1.22

## Ejecución del programa
Para garantizar una fácil ejecución, se creó un `Makefile`, luego en una terminal, el proceso de ejecución va así:
- Para Encriptar: Ejecuta `make run-encrypt in={test_in} out={test_out}`. *test_in* es la carpeta donde se ubican los archivos a encriptar y *test_out* es la carpeta donde se guardan los archivos encriptados.
- Para Desncriptar: Ejecuta `make run-decrypt in={test_in} out={test_out}`. *test_in* es la carpeta donde se ubican los archivos a desencriptar y *test_out* es la carpeta donde se guardan los archivos desencriptados.
- Para Comprimir: Ejecuta `make run-compress in={test_in} out={test_out}`. *test_in* es la carpeta donde se ubican los archivos a comprimir y *test_out* es la carpeta donde se guardan los archivos comprimidos.
- Para Descomprimir: Ejecuta `make run-decompress in={test_in} out={test_out}`. *test_in* es la carpeta donde se ubican los archivos a descomprimir y *test_out* es la carpeta donde se guardan los archivos descomprimir.

*Al utilizar el `Makefile`, no será necesario indicar cuál es el algoritmo que se desea usar.*

Para usar las `flags` para indicar cuál algoritmo desea usar, debe de usar `--comp-alg {Nombre del algoritmo}` para el algoritmo de compresión y `--enc-alg {Nombre del algortimo]` para el algoritmo de encriptado así:
- Para Encriptar: `go run main.go compress.go encrypt.go -e --enc-alg {Nombre del Algoritmo} -i {Ruta de la carpeta de entrada} -o {Ruta de la carpeta de salida}`
- Para Comprimir: `go run main.go compress.go encrypt.go -c --comp-alg {Nombre del Algoritmo} -i {Ruta de la carpeta de entrada} -o {Ruta de la carpeta de salida}`

*Nota: Actualmente solo contamos con el algoritmo de Huffman para compresion (**debe usar en el flag: huff**) y la versión simplificada del AES (**debe usar en el flag: xor**) para encriptado*
