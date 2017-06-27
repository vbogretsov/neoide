package main

/*
#cgo CFLAGS: -Iclangide

#include <errno.h>
#include <string.h>
#include <dlfcn.h>

#include <clang-c/Index.h>

char* errorstr()
{
    return strerror(errno);
}

typedef CXIndex (*clang_create_index_t)(int, int);

CXIndex clang_create_index(
    clang_create_index_t fn, int excludeDeclarationsFromPCH,
    int displayDiagnostics)
{
    return (*fn)(excludeDeclarationsFromPCH, displayDiagnostics);
}

typedef void (*clang_dispose_index_t)(CXIndex);

void clang_dispose_index(clang_dispose_index_t fn, CXIndex index)
{
    (*fn)(index);
}

typedef CXTranslationUnit (*clang_parse_tu_t)(
    CXIndex, const char*, const char* const*, int,
    struct CXUnsavedFile*, unsigned, unsigned);

CXTranslationUnit clang_parse_tu(
    clang_parse_tu_t fn, CXIndex index, const char* source_filename,
    const char* const* command_line_args, unsigned num_command_line_args,
    unsigned options)
{
    return (*fn)(
        index, source_filename, command_line_args, num_command_line_args,
        NULL, 0, options);
}

typedef void (*clang_dispose_tu_t)(CXTranslationUnit);

void clang_dispose_tu(clang_dispose_tu_t fn, CXTranslationUnit tu)
{
    (*fn)(tu);
}

typedef CXCodeCompleteResults* (*clang_complete_at_t)(
    CXTranslationUnit, const char*, unsigned, unsigned, struct CXUnsavedFile*,
    unsigned, unsigned);

CXCodeCompleteResults* complete(
    clang_complete_at_t fn, CXTranslationUnit tu,
    const char* complete_filename, unsigned complete_line,
    unsigned complete_column, struct CXUnsavedFile* unsaved_files,
    unsigned num_unsaved_files, unsigned options)
{
    return (*fn)(
        tu, complete_filename, complete_line, complete_column, unsaved_files,
        num_unsaved_files, options);
}

typedef void (*clang_dispose_completion_t)(CXCodeCompleteResults*);

void clang_dispose_completion(
    clang_dispose_completion_t fn, CXCodeCompleteResults* cr)
{
    (*fn)(cr);
}

typedef const char* (*clang_get_cstring_t)(CXString);

const char* clang_get_stirng(clang_get_cstring_t fn, CXString str)
{
    return (*fn)(str);
}

typedef void (*clang_dispose_string_t)(CXString);

void clang_dispose_string(clang_dispose_string_t fn, CXString str)
{
    (*fn)(str);
}

typedef unsigned (*clang_get_num_completion_chunks_t)(CXCompletionString);

unsigned clang_get_num_completion_chunks(
    clang_get_num_completion_chunks_t fn, CXCompletionString str)
{
    return (*fn)(str);
}

typedef CXString (*clang_get_completion_chunk_text_t)(
    CXCompletionString, unsigned);

CXString clang_get_completion_chunk_text(
    clang_get_completion_chunk_text_t fn,
    CXCompletionString completion_string, unsigned chunk_number)
{
    return (*fn)(completion_string, chunk_number);
}

typedef enum CXCompletionChunkKind (*clang_get_completion_chunk_kind_t)(
    CXCompletionString, unsigned);

enum CXCompletionChunkKind clang_get_completion_chunk_kind(
    clang_get_completion_chunk_kind_t fn, CXCompletionString completion_string,
    unsigned chunk_number)
{
    return (*fn)(completion_string, chunk_number);
}
*/
import "C"

import (
    "errors"
    "fmt"
    "os"
    "unsafe"
)

const (
    IMPORT_ERROR = "unable to import %s"
)

type Clang struct {
    libHandle    unsafe.Pointer
    createIndex  C.clang_create_index_t
    disposeIndex C.clang_dispose_index_t
    parseTu      C.clang_parse_tu_t
    disposeTu    C.clang_dispose_tu_t
}

func ImportError(fname string) error {
    return errors.New(fmt.Sprintf(IMPORT_ERROR, fname))
}

func New(path string) (*Clang, error) {
    handle := C.dlopen(C.CString(path), C.RTLD_NOW|C.RTLD_LOCAL)

    if handle == nil {
        return nil, errors.New(C.GoString(C.errorstr()))
    }

    createIndex := (C.clang_create_index_t)(C.dlsym(
        handle, C.CString("clang_createIndex")))
    if createIndex == nil {
        return nil, ImportError("clang_createIndex")
    }

    disposeIndex := (C.clang_dispose_index_t)(C.dlsym(
        handle, C.CString("clang_disposeIndex")))
    if disposeIndex == nil {
        return nil, ImportError("clang_disposeIndex")
    }

    parseTu := (C.clang_parse_tu_t)(C.dlsym(
        handle, C.CString("clang_parseTranslationUnit")))
    if parseTu == nil {
        return nil, ImportError("clang_parseTranslationUnit")
    }

    disposeTu := (C.clang_dispose_tu_t)(C.dlsym(
        handle, C.CString("clang_disposeTranslationUnit")))
    if disposeTu == nil {
        return nil, ImportError("clang_disposeTranslationUnit")
    }

    complete := (C.clang_complete_at_t)(C.dlsym(
        handle, C.CString("clang_codeCompleteAt")))
    if complete == nil {
        return nil, ImportError("clang_codeCompleteAt")
    }

    disposeCompletion := (C.clang_dispose_completion_t)(C.dlsym(
        handle, C.CString("clang_disposeCodeCompleteResults")))
    if disposeCompletion == nil {
        return nil, ImportError("clang_disposeCodeCompleteResults")
    }

    getString := (C.clang_get_cstring_t)(C.dlsym(
        handle, C.CString("clang_disposeString")))
    if getString == nil {
        return nil, ImportError("clang_disposeString")
    }

    disposeString := (C.clang_dispose_string_t)(C.dlsym(
        handle, C.CString("clang_disposeString")))
    if disposeString == nil {
        return nil, ImportError("clang_disposeString")
    }

    getNumCompletionChunks := (C.clang_get_num_completion_chunks_t)(C.dlsym(
        handle, C.CString("clang_getNumCompletionChunks")))
    if getNumCompletionChunks == nil {
        return nil, ImportError("clang_getNumCompletionChunks")
    }

    getCompletionChunkKind := (C.clang_get_completion_chunk_kind_t)(C.dlsym(
        handle, C.CString("clang_getCompletionChunkKind")))
    if getCompletionChunkKind == nil {
        return nil, ImportError("clang_getCompletionChunkKind")
    }

    getCompletionChunkText := (C.clang_get_completion_chunk_text_t)(C.dlsym(
        handle, C.CString("clang_getCompletionChunkText")))
    if getCompletionChunkText == nil {
        return nil, ImportError("clang_getCompletionChunkText")
    }

    clang := &Clang{
        libHandle: handle, createIndex: createIndex,
        disposeIndex: disposeIndex, parseTu: parseTu, disposeTu: disposeTu}
    return clang, nil
}

func (clang *Clang) CreateIndex(
    excludeDeclarationsFromPCH int, displayDiagnostics int) C.CXIndex {
    return C.clang_create_index(
        clang.createIndex, C.int(excludeDeclarationsFromPCH),
        C.int(displayDiagnostics))
}

func (clang *Clang) DisposeIndex(index C.CXIndex) {
    C.clang_dispose_index(clang.disposeIndex, index)
}

func (clang *Clang) ParseTu(
    index C.CXIndex, filename string) C.CXTranslationUnit {
    return C.clang_parse_tu(
        clang.parseTu, index, C.CString(filename), nil, C.uint(0), C.uint(0))
}

func (clang *Clang) DisposeTu(tu C.CXTranslationUnit) {
    C.clang_dispose_tu(clang.disposeTu, tu)
}

func Start(libclang_path string, source_path string, line int, column int) {
    clang, err := New(libclang_path)

    if err != nil {
        fmt.Println(err)
        os.Exit(1)
    }

    index := clang.CreateIndex(1, 1)
    defer clang.DisposeIndex(index)

    tu := clang.ParseTu(index, source_path)
    defer clang.DisposeTu(tu)

    if tu == nil {
        fmt.Println("parse failed")
    }

    fmt.Println("loaded:", clang)
    fmt.Println("index: ", index)
    fmt.Println("tu:    ", tu)
}
