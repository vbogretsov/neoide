package main

import (
    "fmt"
    "os"
    "strconv"

    "./libclang"
)


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
    libclang.Start(os.Args[1], os.Args[2], line, column)
}