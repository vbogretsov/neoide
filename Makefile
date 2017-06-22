CC				= 	clang
LD 				= 	clang

WARNINGS		=	-Werror

CCFLAGS			=	-std=c11 $(WARNINGS) -c -Ineoided/clangide
LDFLAGS			=	-dynamiclib

ifeq ($(OS),Windows_NT)
	VS_ROOT			=	"c:\Program Files (x86)\Microsoft Visual Studio 14.0"
	WIN_SDK_ROOT	=	"c:\Program Files (x86)\Windows Kits\10"
	WIN_SDK_VERSION	=	10.0.14393.0

	STD_HEADERS		=	-I$(VS_ROOT)\VC\include \
						-I$(WIN_SDK_ROOT)\Include\$(WIN_SDK_VERSION)\ucrt \
						-I$(WIN_SDK_ROOT)\Include\$(WIN_SDK_VERSION)\um \
						-I$(WIN_SDK_ROOT)\Include\$(WIN_SDK_VERSION)\shared
	STD_LIBS		=	-L$(WIN_SDK_ROOT)\Lib\$(WIN_SDK_VERSION)\ucrt\x64 \
						-L$(WIN_SDK_ROOT)\Lib\$(WIN_SDK_VERSION)\um\x64 \
						-L$(VS_ROOT)\VC\lib\amd64
	LIBNAME			=	libclangide.dll
else
	LIBNAME			=	libclangide.dylib
endif

CCFLAGS += $(STD_HEADERS)
CCFLAGS += $(CLANG_HEADERS)

LDFLAGS += $(STD_LIBS)

SRC = neoided/clangide
OBJ = obj
BIN = bin

SRC_FILES := $(wildcard $(SRC)/*.c)
OBJ_FILES := $(addprefix $(OBJ)/,$(notdir $(SRC_FILES:.c=.o)))

GOSRC = neoided


default: $(LIBNAME)
	cp $(BIN)/$(LIBNAME) ./$(GOSRC)/
	cd $(GOSRC) && go build && cd ..
	mv $(GOSRC)/$(GOSRC) $(BIN)/
	@echo done


clean:
	@echo cleaning...
	rm -rf $(OBJ)
	rm -rf $(BIN)

$(OBJ):
	mkdir $(OBJ)
	mkdir $(BIN)


$(OBJ)/%.o: $(SRC)/%.c $(OBJ)
	$(CC) $(CCFLAGS) -c $< -o $@

$(LIBNAME): $(OBJ_FILES)
	$(LD) $(LDFLAGS) -o $(BIN)/$@ $^
