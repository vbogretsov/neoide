package main

import (
    "os"
    "log"

    "github.com/neovim/go-client/nvim/plugin"
    "github.com/neovim/go-client/nvim"

    "./clangide"
    "./types"
)

const (
    LIBCLANG = "/Applications/Xcode.app/Contents/Developer/Toolchains/XcodeDefault.xctoolchain/usr/lib/libclang.dylib"
)

var (
    LOG *log.Logger
)

type Plugin interface {
    Close()

    Enter(path string, action func())
    Save(path string, action func())
    Leave(path string, action func())

    CanComplete(line string) int
    GetCompletions(content string, location *types.Location) *[]map[string]string

    FindDefenition(content string, location *types.Location) *[]types.Location
    FindDeclaration(content string, location *types.Location) *[]types.Location
    FindReferences(content string, location *types.Location) *[]types.Location
    FindAssingments(content string, location *types.Location) *[]types.Location
}

func CreateClangIde() (Plugin, error) {
    // TODO: receive form settings
    libclang := LIBCLANG
    flags := []string{}

    return clangide.New(libclang, flags)
}

// TODO: load from shared libraries
var Plugins = map[string]func()(Plugin, error){
    "c": CreateClangIde, "cpp": CreateClangIde}

type NeoIde struct {
    plugins     map[string]Plugin
    completions map[string]*map[string]string
}

func New() *NeoIde {
    plugins := make(map[string]Plugin)
    return &NeoIde{plugins: plugins}
}

func (ide *NeoIde) Close() {
    for _, plug := range ide.plugins {
        plug.Close()
    }
}

func (ide *NeoIde) Configure(vim *nvim.Nvim, args []interface{}) error {
    LOG.Println("configured", args)
    return nil
}

func (ide *NeoIde) Enter(vim *nvim.Nvim, filetype string, path string) {
    if _, ok := ide.plugins[filetype]; !ok {
        if functor, ok := Plugins[filetype]; ok {
            plug, err := functor()
            if err == nil {
                ide.plugins[filetype] = plug
            } else {
                // TODO: log error
            }
        }
    }
    if plug, ok := ide.plugins[filetype]; !ok {
        plug.Enter(path, func(){})
    }
}

func (ide *NeoIde) Save(vim *nvim.Nvim, filetype string, path string) {
    if plug, ok := ide.plugins[filetype]; ok {
        plug.Save(path, func(){})
    }
}

func (ide *NeoIde) Leave(vim *nvim.Nvim, filetype string, path string) {
    if plug, ok := ide.plugins[filetype]; ok {
        plug.Leave(path, func(){})
    }
}

func (ide *NeoIde) FindCompletions(
    filetype string, content string,
    path string, line int) (*map[string]interface{}, error) {

    if plug, ok := ide.plugins[filetype]; ok {
        // if match
        // get completions
        // else return completions
        column := plug.CanComplete("")
        if column > 0 {
            location := &types.Location{path, line, column}
            completions := plug.GetCompletions(content, location)
            result := map[string]interface{}{"words": &completions, "position": column}
            return &result, nil
        }

        return nil, nil
    }

    return nil, nil
}

func (ide *NeoIde) FindDefenition(
    filetype string, content string, path string,
    line int, column int) *[]types.Location {

    if plug, ok := ide.plugins[filetype]; ok {
        location := &types.Location{path, line, column}
        return plug.FindDefenition(content, location)
    }

    return nil
}

func (ide *NeoIde) FindDeclaration(
    filetype string, content string, path string,
    line int, column int) *[]types.Location {

    if plug, ok := ide.plugins[filetype]; ok {
        location := &types.Location{path, line, column}
        return plug.FindDeclaration(content, location)
    }

    return nil
}

func (ide *NeoIde) FindReferences(
    filetype string, content string, path string,
    line int, column int) *[]types.Location {

    if plug, ok := ide.plugins[filetype]; ok {
        location := &types.Location{path, line, column}
        return plug.FindReferences(content, location)
    }

    return nil
}

func (ide *NeoIde) FindAssingments(
    filetype string, content string, path string,
    line int, column int) *[]types.Location {

    if plug, ok := ide.plugins[filetype]; ok {
        location := &types.Location{path, line, column}
        return plug.FindAssingments(content, location)
    }

    return nil
}

func main() {
    flags := os.O_APPEND | os.O_WRONLY | os.O_CREATE
    file, err := os.OpenFile("/tmp/neoide.log", flags, 0666)
    if err != nil {
        panic(err)
    }
    defer file.Close()
    LOG = log.New(file, "", log.LstdFlags | log.Lshortfile)
    LOG.Println("neoide started")

    neoide := New()

    plugin.Main(func(p *plugin.Plugin) error {
        p.HandleFunction(
            &plugin.FunctionOptions{Name: "_neoide_configure"}, neoide.Configure)
        // p.HandleFunction(
        //     &plugin.FunctionOptions{Name: "NeoIdeEnter"}, neoide.Enter)
        // p.HandleFunction(
        //     &plugin.FunctionOptions{Name: "NeoIdeSave"}, neoide.Save)
        // p.HandleFunction(
        //     &plugin.FunctionOptions{Name: "NeoIdeLeave"}, neoide.Leave)
        // p.HandleFunction(
        //     &plugin.FunctionOptions{Name: "NeoIdeFindCompletions"}, neoide.FindCompletions)
        // p.HandleFunction(
        //     &plugin.FunctionOptions{Name: "NeoIdeFindDefenition"}, neoide.FindDefenition)
        // p.HandleFunction(
        //     &plugin.FunctionOptions{Name: "NeoIdeFindDeclaration"}, neoide.FindDeclaration)
        // p.HandleFunction(
        //     &plugin.FunctionOptions{Name: "NeoIdeFindReferences"}, neoide.FindReferences)
        // p.HandleFunction(
        //     &plugin.FunctionOptions{Name: "NeoIdeFindAssingments"}, neoide.FindAssingments)
        return nil
    })
}
