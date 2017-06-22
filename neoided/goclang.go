package main

/*
#cgo CFLAGS: -Iclangide/
#cgo LDFLAGS: -L. -lclangide
#include "ide.h"

extern void GoOnCompletion(unsigned, void*, completion_t*);

static void OnCompletion(unsigned index, void* ctx, completion_t* completion) {
    GoOnCompletion(index, ctx, completion);
}

static void completions_foreach(completions_t* completions, void* ctx)
{
    completions_iter(completions, ctx, &OnCompletion);
}

*/
import "C"

import (
    "fmt"
    "io/ioutil"
    "os"
    "strconv"
    "unsafe"
)

// "/Applications/Xcode.app/Contents/Developer/Toolchains/XcodeDefault.xctoolchain/usr/lib/libclang.dylib"

type Completion *C.completion_t

type Completions [5000]map[string]string


//export GoOnCompletion
func GoOnCompletion(index C.uint, ctx unsafe.Pointer, completion Completion) {
    completions := (*[5000]map[string]string)(ctx)
    completions[index] = map[string]string{
        "abbr": C.GoString(&completion.abbr[0]),
        "word": C.GoString(&completion.word[0])}
}


func Start(libclang string, filename string, line int, column int) {
    // TODO: handle errors
    ide := C.ide_alloc(C.CString(libclang))
    defer C.ide_free(ide)

    C.ide_on_file_open(ide, C.CString(filename))

    content, err := ioutil.ReadFile(filename)
    if err != nil {
        fmt.Println(err)
    }

    source := string(content)

    completions := C.ide_find_completions(
        ide,
        C.CString(filename),
        C.uint(line),
        C.uint(column),
        C.CString(source),
        C.uint(len(source)))
    defer C.completions_free(completions)

    results := make([]map[string]string, C.completions_count(completions))

    ctx := unsafe.Pointer(&results[0])
    C.completions_foreach(completions, ctx)

    fmt.Println(results)
}


func main() {
    if len(os.Args) != 5 {
        fmt.Println("usage: neoided <libclang> <filename> <line> <column>")
        os.Exit(1)
    }
    line, err := strconv.Atoi(os.Args[3])
    if err != nil {
        fmt.Println("line should be an integer")
        os.Exit(1)
    }
    column, err := strconv.Atoi(os.Args[4])
    if err != nil {
        fmt.Println("column should be an integer")
        os.Exit(1)
    }
    Start(os.Args[1], os.Args[2], line, column)
}
