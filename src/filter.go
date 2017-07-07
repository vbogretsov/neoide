package main

import (
    "sort"
    "strconv"

    "./types"
)

type byPrio []map[string]string

func (p byPrio) Len() int {
    return len(p)
}

func (p byPrio) Swap(i, j int) {
    p[i], p[j] = p[j], p[i]
}

func (p byPrio) Less(i, j int) bool {
    a, _ := strconv.ParseFloat(p[i]["prio"], 64)
    b, _ := strconv.ParseFloat(p[j]["prio"], 64)
    return a > b
}

func Distance(word string, pattern string) float64 {
    result := 0.0
    lastMatch := 0.0

    if len(pattern) == 0 {
        return 1.0
    }

    for pi := 0; pi < len(pattern); pi++ {
        for wi := 0; wi < len(word); wi++ {
            match := 0.0
            if pattern[pi] == word[wi] {
                match = 1.0 / (lastMatch + 1.0)
            }
            if match != 0 {
                result += match;
                lastMatch = 0;
                break;
            }
            lastMatch += 1.0
        }
    }

    return result / float64(len(pattern))
}

func Filter(completions *[]map[string]string, word string) *[]map[string]string {
    result := []map[string]string{}
    for _, value := range *completions {
        prio := Distance(value["word"], word)
        types.LOG.Printf("[%s, %s] ~ %f\n", value["word"], word, prio)
        if prio > 0.33 {
            value["prio"] = strconv.FormatFloat(prio, 'E', -1, 64)
            result = append(result, value)
        }
    }
    sort.Sort(byPrio(result))
    return &result
}