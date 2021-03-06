package main

import (
    "log"
    "os"
    "github.com/vbogretsov/neoide/src/types"
    "github.com/neovim/go-client/nvim/plugin"
    "github.com/neovim/go-client/nvim"

    "./clangide"
)

func openLogFile(path string) *os.File {
    flags := os.O_APPEND | os.O_WRONLY | os.O_CREATE
    file, err := os.OpenFile(path, flags, 0666)
    if err != nil {
        panic(err)
    }
    return file
}

func loadPlugins() map[string]func(*nvim.Nvim)(types.Plugin, error) {
    plugins := map[string]func(*nvim.Nvim)(types.Plugin, error) {
        "c": clangide.CreateCIde, "cpp": clangide.CreateCppIde}
    return plugins
}

func main() {
    file := openLogFile("/tmp/neoide.log")
    defer file.Close()

    types.LOG = log.New(file, "", log.LstdFlags | log.Lshortfile)
    types.LOG.Println("neoide started")

    neoide := New(loadPlugins())

    plugin.Main(func(p *plugin.Plugin) error {
        p.HandleFunction(
            &plugin.FunctionOptions{Name: "_neoide_bufenter"},
            neoide.Enter)
        p.HandleFunction(
            &plugin.FunctionOptions{Name: "_neoide_bufsave"},
            neoide.Save)
        p.HandleFunction(
            &plugin.FunctionOptions{Name: "_neoide_bufclose"},
            neoide.Leave)
        p.HandleFunction(
            &plugin.FunctionOptions{Name: "_neoide_find_completions"},
            neoide.FindCompletions)
        p.HandleFunction(
            &plugin.FunctionOptions{Name: "_neoide_show_completions"},
            neoide.ShowCompletions)
        p.HandleFunction(
            &plugin.FunctionOptions{Name: "_neoide_get_completions"},
            neoide.GetCompletions)
        p.HandleFunction(
            &plugin.FunctionOptions{Name: "_neoide_find_defenition"},
            neoide.FindDefenition)
        p.HandleFunction(
            &plugin.FunctionOptions{Name: "_neoide_find_declaration"},
            neoide.FindDeclaration)
        p.HandleFunction(
            &plugin.FunctionOptions{Name: "_neoide_find_references"},
            neoide.FindReferences)
        p.HandleFunction(
            &plugin.FunctionOptions{Name: "_neoide_find_assingments"},
            neoide.FindAssingments)
        return nil
    })
}