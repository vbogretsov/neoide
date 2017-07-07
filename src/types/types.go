/**
 * Types definitions for NeoIDE.
 * 
 * Jul 2 2017 Vladimir Bogretsov <bogrecov@gmail.com>
 */
package types

import "log"

var (
    LOG *log.Logger
)

/**
 * Represents location in a file.
 */
type Location struct {
    Path   string
    Line   int
    Column int
}

type Closable interface {
    Close()
}

type FileController interface {
    Enter(path string, action func())
    Save(path string, action func())
    Leave(path string, action func())
}

type CodeCompleter interface {
    CanComplete(line string) int
    GetCompletions(content string, location *Location) *[]map[string]string
}

type CodeNavigator interface {
    FindDefenition(content string, location *Location) *[]Location
    FindDeclaration(content string, location *Location) *[]Location
    FindReferences(content string, location *Location) *[]Location
    FindAssingments(content string, location *Location) *[]Location
}

type Plugin interface {
    Closable
    FileController
    CodeCompleter
    CodeNavigator
}