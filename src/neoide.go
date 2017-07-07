package main

import (
    "errors"
    "math/rand"
    "strings"

    "github.com/neovim/go-client/nvim"

    "./types"
)

type Neoide struct {
    funcs         map[string]func(*nvim.Nvim)(types.Plugin, error)
    plugs         map[string]types.Plugin
    completions   *[]map[string]string
    completion_id int
}

func New(funcs map[string]func(*nvim.Nvim)(types.Plugin, error)) *Neoide {
    plugs := make(map[string]types.Plugin)
    return &Neoide{funcs: funcs, plugs: plugs}
}

func (ide *Neoide) Close() {
    for _, plug := range ide.plugs {
        plug.Close()
    }
}

func (ide *Neoide) Enter(vim *nvim.Nvim, args []interface{}) error {
    filetype, ok := args[0].(string)
    if !ok {
        return errors.New("filetype should be a string")
    }

    path, ok := args[1].(string)
    if !ok {
        return errors.New("path should be a string")
    }

    var err error = nil

    if _, ok := ide.plugs[filetype]; !ok {
        if functor, ok := ide.funcs[filetype]; ok {
            plug, err := functor(vim)
            if err == nil {
                ide.plugs[filetype] = plug
            }
        }
    }

    if plug, ok := ide.plugs[filetype]; ok {
        plug.Enter(path, func(){vim.Call("neoide#info", nil, "file ready")})
    }

    return err
}

func (ide *Neoide) Save(vim *nvim.Nvim, args []interface{}) error {
    filetype, ok := args[0].(string)
    if !ok {
        return errors.New("filetype should be a string")
    }

    path, ok := args[1].(string)
    if !ok {
        return errors.New("path should be a string")
    }

    if plug, ok := ide.plugs[filetype]; ok {
        plug.Save(path, func(){})
    }

    return nil
}

func (ide *Neoide) Leave(vim *nvim.Nvim, args []interface{}) error {
    filetype, ok := args[0].(string)
    if !ok {
        return errors.New("filetype should be a string")
    }

    path, ok := args[1].(string)
    if !ok {
        return errors.New("path should be a string")
    }

    if plug, ok := ide.plugs[filetype]; ok {
        plug.Leave(path, func(){})
    }

    return nil
}

func GatherCompletions(
    vim *nvim.Nvim, column int, plug types.Plugin) *[]map[string]string {

    batch := vim.NewBatch()

    var path string
    var content []string
    var line int

    batch.Call("expand", &path, "%:p")
    batch.Call("getline", &content, 1, "$")
    batch.Call("line", &line, ".")
    err := batch.Execute()

    var completions *[]map[string]string
    if err == nil {
        location := &types.Location{path, line, column}
        text := strings.Join(content, "\n")
        completions = plug.GetCompletions(text, location)
    } else {
        completions = &[]map[string]string{}
        vim.Call("neoide#error", nil, err)
    }

    return completions
}

func (ide *Neoide) GetCompletions(
    vim *nvim.Nvim, args []interface{}) (*[]map[string]string, error) {

    if ide.completions == nil {
        return &[]map[string]string{}, nil
    }

    word, ok := args[0].(string)
    if !ok {
        word = ""
    }
    word = strings.TrimSpace(word)

    if word == "" {
        return ide.completions, nil
    }

    result := Filter(ide.completions, word)

    return result, nil
}

func (ide *Neoide) ShowCompletions(vim *nvim.Nvim, args []interface{}) {
    filetype, ok := args[0].(string)
    if !ok {
        vim.Call("neoide#error", nil, "filetype should be a string")
        return
    }

    column, ok := args[1].(int64)
    if !ok {
        vim.Call("neoide#error", nil, "column should be an integer")
        return
    }

    if plug, ok := ide.plugs[filetype]; ok {
        ide.completions = GatherCompletions(vim, int(column), plug)
        vim.Call("neoide#show_popup", nil, column - 1)
    }
}

func (ide *Neoide) FindCompletions(vim *nvim.Nvim, args []interface{}) {

    filetype, ok := args[0].(string)
    if !ok {
        vim.Call("neoide#error", nil, "filetype should be a string")
        return
    }

    line, ok := args[1].(string)
    if !ok {
        vim.Call("neoide#error", nil, "line should be a string")
        return
    }

    plug, ok := ide.plugs[filetype]
    if !ok {
        return
    }

    column := plug.CanComplete(line)

    if column > 0 {
        completion_id := rand.Int()
        ide.completion_id = completion_id
        ide.completions = GatherCompletions(vim, column, plug)

        if completion_id == ide.completion_id {
            vim.Call("neoide#show_popup", nil, column - 1)
        }
    }
}

func (ide *Neoide) FindDefenition(
    filetype string, content string, path string,
    line int, column int) *[]types.Location {

    if plug, ok := ide.plugs[filetype]; ok {
        location := &types.Location{path, line, column}
        return plug.FindDefenition(content, location)
    }

    return nil
}

func (ide *Neoide) FindDeclaration(
    filetype string, content string, path string,
    line int, column int) *[]types.Location {

    if plug, ok := ide.plugs[filetype]; ok {
        location := &types.Location{path, line, column}
        return plug.FindDeclaration(content, location)
    }

    return nil
}

func (ide *Neoide) FindReferences(
    filetype string, content string, path string,
    line int, column int) *[]types.Location {

    if plug, ok := ide.plugs[filetype]; ok {
        location := &types.Location{path, line, column}
        return plug.FindReferences(content, location)
    }

    return nil
}

func (ide *Neoide) FindAssingments(
    filetype string, content string, path string,
    line int, column int) *[]types.Location {

    if plug, ok := ide.plugs[filetype]; ok {
        location := &types.Location{path, line, column}
        return plug.FindAssingments(content, location)
    }

    return nil
}
