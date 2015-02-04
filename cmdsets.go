package main

import (
    "fmt"
    "strings"
    "flag"

    "github.com/lmika-bom/oaipmh/client"
)


// ---------------------------------------------------------------------------------------------------
// Sets commands

type SetsCommand struct {
    Ctx             *Context
    flagDetailed    *bool
}

func (lc *SetsCommand) Flags(fs *flag.FlagSet) *flag.FlagSet {
    lc.flagDetailed = fs.Bool("l", false, "Use detailed listing format")
    return fs
}

func (lc *SetsCommand) Run(args []string) {
    lc.Ctx.Session.ListSets(0, -1, func(res oaipmh.OaipmhSet) bool {
        if (*(lc.flagDetailed)) {
            fmt.Printf("Name: %s\nSpec: %s\n\n", res.Spec, res.Name)
            descrLines := strings.Split(res.Descr.OaiDC.Descr, "\n")
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
