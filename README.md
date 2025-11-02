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

## Funcionalidades Pendientes:
- Se debe desarrollar un algoritmo para compresión y descompresión de archivos.
