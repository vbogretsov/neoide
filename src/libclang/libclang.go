/**
 * Simplified Go bindings for clang C API.
 *
 * Jun 30 2017 Vladimir Bogretsov <bogrecov@gmail.com>
 */
package libclang

/*
#include <stdlib.h>
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
    "fmt"
    "unsafe"
)

const (
    TUPrecompiledPreamble = C.CXTranslationUnit_PrecompiledPreamble
    TUCacheCompletionResults = C.CXTranslationUnit_CacheCompletionResults
    TUIncomplete = C.CXTranslationUnit_Incomplete
)

const (
    CCIncludeMacros = C.CXCodeComplete_IncludeMacros
    CCIncludeBriefComments = C.CXCodeComplete_IncludeBriefComments
    CCIncludeCodePatterns = C.CXCodeComplete_IncludeCodePatterns
)

type CStrings struct {
    array **C.char
    size  C.uint
}

//export ReadCompletion
func ReadCompletion(
    completion *C.completion_t, index C.uint, ctx unsafe.Pointer) {

    i := int(index)
    completions := (*[1 << 30]map[string]string)(ctx)
    completions[i] = map[string]string{
        "abbr": C.GoString(&completion.abbr[0]),
        "word": C.GoString(&completion.word[0]),
        "menu": "[clang]"}
}

func ToCStrings(array []string) *CStrings {
    result := C.make_string_array(C.uint(len(array)))
    ptr := (*[1 << 30]*C.char)(unsafe.Pointer(result))

    for i := 0; i < len(array); i++ {
        ptr[i] = C.CString(array[i])
    }

    return &CStrings{result, C.uint(len(array))}
}

func (strings *CStrings) Free() {
    C.free_string_array(strings.array, strings.size)
}

func ClangError() string {
    return C.GoString(C.libclang_error())
}

type Clang struct {
    handle *C.libclang_t
}

type Index struct {
    handle C.index_t
}

type TranslationUnit struct {
    handle C.translation_unit_t
}

func Load(sopath string) (*Clang, error) {
    handle := C.libclang_load(C.CString(sopath))
    if handle == nil {
        return nil, errors.New(fmt.Sprintf(
            "unable to load libclang: %s", C.GoString(C.libclang_error())))
    }
    return &Clang{handle: handle}, nil
}

func (clang *Clang) Close() {
    C.libclang_free(clang.handle)
}

func (clang *Clang) CreateIndex(
    excludeDeclarationsFromPCH int, displayDiagnostics int) *Index {

    handle := C.libclang_create_index(
        clang.handle, C.int(excludeDeclarationsFromPCH),
        C.int(displayDiagnostics))
    return &Index{handle: handle}
}

func (clang *Clang) CloseIndex(index *Index) {
    C.libclang_dispose_index(clang.handle, index.handle)
}

// TODO: add errors handling
func (clang *Clang) ParseTu(
    index *Index, filename string,
    flags *CStrings, options int) *TranslationUnit {

    handle := C.libclang_parse_tu(
        clang.handle, index.handle, C.CString(filename),
        flags.array, flags.size, C.uint(options))
    return &TranslationUnit{handle: handle}
}

func (clang *Clang) ReparseTu(tu *TranslationUnit, options int) {
    C.libclang_reparse_tu(clang.handle, tu.handle, C.uint(options))
}

func (clang *Clang) CloseTu(tu *TranslationUnit) {
    C.libclang_dispose_tu(clang.handle, tu.handle)
}

// TODO: add error handling
func (clang *Clang) Complete(
    tu *TranslationUnit, options int, content string, filename string,
    line int, column int) *[]map[string]string {

    results := C.libclang_complete_at(
        clang.handle, tu.handle, C.uint(options), C.CString(filename),
        C.CString(content), C.uint(len(content)), C.uint(line), C.uint(column))
    defer C.libclang_completions_free(clang.handle, results)

    if results == nil || results.NumResults == 0 {
        return &[]map[string]string{}
    }

    completions := make([]map[string]string, results.NumResults)
    ctx := unsafe.Pointer(&completions[0])
    C.copy_completions(clang.handle, results, ctx)

    return &completions
}