SRC = src
BIN = bin
EXE = $(BIN)/neoided

SRC_LIBCLANG := $(wildcard $(SRC)/libclang/*.go)
SRC_CLANGIDE := $(wildcard $(SRC)/clangide/*.go)
SRC_MAIN     := $(wildcard $(SRC)/*.go)

SOURCES = $(SRC_LIBCLANG) $(SRC_CLANGIDE) $(SRC_MAIN)

default: $(EXE)
	@echo done

clean:
	rm -rf $(BIN)

dependencies:
	go get "github.com/neovim/go-client/nvim"
	go get "github.com/neovim/go-client/nvim/plugin"
	go get "github.com/vbogretsov/neoide/src/libclang"
	go get "github.com/vbogretsov/neoide/src/types"

$(BIN):
	mkdir -p $(BIN)

$(EXE): $(SOURCES) $(BIN) dependencies
	go build -o $(EXE) ./$(SRC)