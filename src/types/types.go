/**
 * Types definitions for NeoIDE.
 * 
 * Jul 2 2017 Vladimir Bogretsov <bogrecov@gmail.com>
 */
package types

/**
 * Represents location in a file.
 */
type Location struct {
    Path   string
    Line   int
    Column int
}