package libclang

/*
#cgo CFLAGS: -Iinterop
#cgo LDFLAGS: -L${SRCDIR} -lclanginterop
#include <libclang.h>
*/
import "C"

func Start(libclangPath string, sourcePath string, line int, column int) {
    C.test()
}