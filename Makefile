# Nombre del módulo
MODULE = kryptr


# Ejecutables
ENCRYPT_BIN = encrypt
MAIN_BIN = main

# Comandos
GO = go

all: build

# Compilar ambos binarios
build: $(ENCRYPT_BIN) $(MAIN_BIN)

$(ENCRYPT_BIN): encrypt.go
	$(GO) build -o $(ENCRYPT_BIN)

$(MAIN_BIN): main.go
	$(GO) build -o $(MAIN_BIN)

# Ejecutar encrypt
run-encrypt: $(ENCRYPT_BIN) $(MAIN_BIN)
	./$(MAIN_BIN) -e -i $(in) -o $(out)

# Ejecutar decrypt
run-decrypt: $(ENCRYPT_BIN) $(MAIN_BIN)
	./$(MAIN_BIN) -u -i $(in) -o $(out)

# Limpiar ejecutables
clean:
	rm -f $(ENCRYPT_BIN) $(MAIN_BIN)