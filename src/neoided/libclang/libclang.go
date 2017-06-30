package libclang

/*
#cgo CFLAGS: -Iinterop
#cgo LDFLAGS: -L${SRCDIR} -L${SRCDIR}/../../../bin -lclanginterop
#include <libclang.h>

extern void ReadCompletion(completion_t*, unsigned, void*);

static void copy_completions(
    libclang_t* so, completion_results_t* completions, void* ctx)
{
    libclang_completions_foreach(so, completions, ctx, &ReadCompletion);
}
*/
import "C"

import (
    "errors"
    "io/ioutil"
    "fmt"
    "os"
    "unsafe"
)

const ParseOptions =
    C.CXTranslationUnit_PrecompiledPreamble |
    C.CXTranslationUnit_CacheCompletionResults |
    C.CXTranslationUnit_Incomplete

//export ReadCompletion
func ReadCompletion(
    completion *C.completion_t, index C.uint, ctx unsafe.Pointer) {

    i := int(index)
    completions := (*[1 << 30]map[string]string)(ctx)
    completions[i] = map[string]string{
        "abbr": C.GoString(&completion.abbr[0]),
        "word": C.GoString(&completion.word[0])}
}

func Load(sopath string) (*C.libclang_t, error) {
    clang := C.libclang_load(C.CString(sopath))
    if clang == nil {
        return nil, errors.New(C.GoString(C.libclang_error()))
    }
    return clang, nil
}

func (clang *C.libclang_t) Close() {
    if clang != nil {
        C.libclang_free(clang)
    }
}

func (clang *C.libclang_t) CreateIndex(
    excludeDeclarationsFromPCH int, displayDiagnostics int) C.index_t {

    index := C.libclang_create_index(
        clang, C.int(excludeDeclarationsFromPCH), C.int(displayDiagnostics))
    return index
}

func (clang *C.libclang_t) CloseIndex(index C.index_t) {
    C.libclang_dispose_index(clang, index)
}

// TODO: add errors handling
// TODO: pass flags
func (clang *C.libclang_t) ParseTu(
    index C.index_t, filename string,
    flags **C.char, num_flags int, options int) C.translation_unit_t {

    tu := C.libclang_parse_tu(
        clang, index, C.CString(filename), flags,
        C.uint(num_flags), C.uint(options))
    return tu
}

func (clang *C.libclang_t) CloseTu(tu C.translation_unit_t) {
    C.libclang_dispose_tu(clang, tu)
}

// TODO: add error handling
func (clang *C.libclang_t) Complete(
    tu C.translation_unit_t, options int, filename string,
    content string, line int, column int) *[]map[string]string {

    results := C.libclang_complete_at(
        clang, tu, C.uint(options), C.CString(filename), C.CString(content),
        C.uint(len(content)), C.uint(line), C.uint(column))
    defer C.libclang_completions_free(clang, results)

    completions := make([]map[string]string, results.NumResults)
    ctx := unsafe.Pointer(&completions[0])
    C.copy_completions(clang, results, ctx)

    return &completions
}

func Start(libclangPath string, sourcePath string, line int, column int) {
    clang, err := Load(libclangPath)
    defer clang.Close()

    if clang == nil {
        fmt.Println(C.GoString(C.libclang_error()))
        os.Exit(1)
    }

    index := clang.CreateIndex(1, 1)
    defer clang.CloseIndex(index)

    tu := clang.ParseTu(index, sourcePath, nil, 0, ParseOptions)
    defer clang.CloseTu(tu)

    if tu == nil {
        fmt.Println("parse failed")
        os.Exit(1)
    }

    file, err := ioutil.ReadFile(sourcePath)
    if err != nil {
        fmt.Println(err)
        os.Exit(1)
    }
    content := string(file)

    completions := clang.Complete(tu, 0, sourcePath, content, line, column)

    fmt.Println("total:", len(*completions))

    for _, completion := range (*completions) {
        fmt.Println(completion)
    }

    fmt.Println("done")
}