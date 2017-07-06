/**
 * Psugins interfaces for NEOIDE.
 * 
 * Jul 4 2017 Vladimir Bogretsov <bogrecov@gmail.com>
 */
package plugins

import (
    "github.com/neovim/go-client/nvim"

    "../types"
    "../clangide"
)

/**
 * Plugins list. Should be removed once plugins will be available for OS X.
 */
var PLUGINS = map[string]func(*nvim.Nvim)(types.Plugin, error){
    "c": clangide.CreateCIde, "cpp": clangide.CreateCppIde}