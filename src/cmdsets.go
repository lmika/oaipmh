package main

import (
    "fmt"
    "strings"
    "flag"
)


// ---------------------------------------------------------------------------------------------------
// Sets commands

type SetsCommand struct {
    Ctx             *Context
    flagDetailed    *bool
}

func (lc *SetsCommand) Flags(fs *flag.FlagSet) *flag.FlagSet {
    lc.flagDetailed = fs.Bool("l", false, "List sets in detail.")
    return fs
}

func (lc *SetsCommand) Run(args []string) {
    lc.Ctx.Session.ListSets(0, -1, func(res ListSetResult) bool {
        if (*(lc.flagDetailed)) {
            fmt.Printf("Name: %s\nSpec: %s\n\n", res.Spec, res.Name)
            descrLines := strings.Split(res.Description, "\n")
            for _, v := range descrLines {
                v = strings.Trim(v, " \n\t")
                if (v != "") {
                    fmt.Printf("%s\n", v)
                }
            }
            fmt.Println("---")
        } else {
            fmt.Printf("%s\n", res.Name)
        }
        return true
    })
}
