STD				=	-std=c11
WARNINGS		=	-Werror

CCFLAGS			=	$(STD) $(WARNINGS) -c
LDFLAGS			=	-dynamiclib

SRC				=	./src/neoided
OBJ				=	obj
BIN				=	bin

NEOIDE			=	neoided

LIBCLANG_SRC		=	$(SRC)/libclang/
LIBCLANG_INTEROP	=	$(LIBCLANG_SRC)/interop
LIBCLANG_CCFLAGS	=	-I$(LIBCLANG_INTEROP)

LIBCLANG_SRC_FILES := $(wildcard $(LIBCLANG_INTEROP)/*.c)
LIBCLANG_OBJ_FILES := $(addprefix $(OBJ)/,$(notdir $(LIBCLANG_SRC_FILES:.c=.o)))
LIBCLANG_BIN = libclanginterop.dylib

GOLIBCLANG_SRC := $(wildcard $(LIBCLANG_SRC)/*.go)

default: $(BIN)/$(NEOIDE)
	@echo done

clean:
	@echo cleaning...
	rm -rf $(OBJ)
	rm -rf $(BIN)

$(OBJ):
	mkdir $(OBJ)
	mkdir $(BIN)

$(OBJ)/%.o: $(LIBCLANG_INTEROP)/%.c $(OBJ)
	$(CC) $(CCFLAGS) $(LIBCLANG_CCFLAGS) $< -o $@

$(BIN)/$(LIBCLANG_BIN): $(LIBCLANG_OBJ_FILES)
	$(CC) $(LDFLAGS) -o $@ $^


$(BIN)/$(NEOIDE): $(GOLIBCLANG_SRC) $(BIN)/$(LIBCLANG_BIN)
	# cp $(BIN)/$(LIBCLANG_BIN) ./
	go build -o $(BIN)/$(NEOIDE) $(SRC)
