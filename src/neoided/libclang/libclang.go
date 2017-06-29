package libclang

/*
#cgo CFLAGS: -Iinterop
#cgo LDFLAGS: -L${SRCDIR} -L${SRCDIR}/../../../bin -lclanginterop
#include <libclang.h>
*/
import "C"

import "fmt"
import "os"

const ParseOptions = C.CXTranslationUnit_PrecompiledPreamble | C.CXTranslationUnit_CacheCompletionResults | C.CXTranslationUnit_Incomplete

func Start(libclangPath string, sourcePath string, line int, column int) {
    libclang := C.libclang_load(C.CString(libclangPath))
    defer C.libclang_free(libclang)

    if libclang == nil {
        fmt.Println(C.GoString(C.libclang_error()))
        os.Exit(1)
    }

    index := C.libclang_create_index(libclang, C.int(1), C.int(1))
    defer C.libclang_dispose_index(libclang, index)

    tu := C.libclang_parse_tu(
        libclang, index, C.CString(sourcePath), nil, C.uint(0), C.uint(0))

    if tu == nil {
        fmt.Println("parse failed")
        os.Exit(1)
    }

    defer C.libclang_dispose_tu(libclang, tu)

    fmt.Println("done")
}