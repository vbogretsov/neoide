/**
 * IDE functions for C/C++ based on clang C API.
 *
 * Jun 30 2017 Vladimir Bogretsov <bogrecov@gmail.com>
 */
package clangide

import (
    "io/ioutil"
    "fmt"
    "os"

    // TODO: avoid path dependent imports
    "../types"
    "../libclang"
)

const ParseOptions =
    libclang.TUPrecompiledPreamble |
    libclang.TUCacheCompletionResults |
    libclang.TUIncomplete

const CompleteOptions = libclang.CCIncludeMacros

type Ide struct {
    clang *libclang.Clang
    index *libclang.Index
    flags *libclang.CStrings
    units map[string]*libclang.TranslationUnit
}

func New(sopath string, flags []string) (*Ide, error) {
    clang, err := libclang.Load(sopath)

    if err != nil {
        return nil, err
    }

    index := clang.CreateIndex(1, 1)
    units := make(map[string]*libclang.TranslationUnit)
    array := libclang.ToCStrings(flags)

    return &Ide{clang: clang, flags: array, index: index, units: units}, nil
}

func (ide *Ide) Close() {
    for _, tu := range ide.units {
        ide.clang.CloseTu(tu)
    }
    ide.flags.Free()
    ide.clang.Close()
}

func (ide *Ide) OpenFile(path string, action func()) {
    if old, ok := ide.units[path]; ok {
        ide.clang.CloseTu(old)
    }

    ide.units[path] = ide.clang.ParseTu(
        ide.index, path, ide.flags, ParseOptions)
    action()
}

func (ide *Ide) SaveFile(path string, action func()) {
    if tu, ok := ide.units[path]; ok {
        ide.clang.ReparseTu(tu, ParseOptions)
        action()
    }
}

func (ide *Ide) CloseFile(path string, action func()) {
    if tu, ok := ide.units[path]; ok {
        ide.clang.CloseTu(tu)
        action()
    }
}

func (ide *Ide) FindCompletions(
    content string, location *types.Location) *[]map[string]string {

    var completions *[]map[string]string = nil

    if tu, ok := ide.units[location.Path]; ok {
        completions = ide.clang.Complete(
            tu, CompleteOptions, content,
            location.Path, location.Line, location.Column)
    }

    return completions
}

func (ide *Ide) FindDefenition(
    content string, location *types.Location) *[]types.Location {

    return nil
}

func (ide *Ide) FindDeclaration(
    content string, location *types.Location) *[]types.Location {

    return nil
}

func (ide *Ide) FindReferences(
    content string, location *types.Location) *[]types.Location {

    return nil
}

func (ide *Ide) FindAssingments(
    content string, location *types.Location) *[]types.Location {

    return nil
}

func Start(libclangPath string, sourcePath string, line int, column int) {
    flags := []string{"-I/usr/include", "-I~/ports/include"}

    ide, err := New(libclangPath, flags)

    if err == nil {
        defer ide.Close()
    } else {
        fmt.Println(err)
        os.Exit(1)
    }

    ide.OpenFile(sourcePath, func(){})

    file, err := ioutil.ReadFile(sourcePath)

    if err != nil {
        fmt.Println(err)
        os.Exit(1)
    }

    completions := ide.FindCompletions(
        string(file), &types.Location{sourcePath, line, column})

    for _, completion := range (*completions) {
        fmt.Println(completion)
    }

    fmt.Println("done")
}