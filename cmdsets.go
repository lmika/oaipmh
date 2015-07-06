package main

import (
    "fmt"
    "strings"
    "flag"
    "bufio"

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
            fmt.Printf("%s: %s\n", res.Spec, res.Name)
            descrLines := strings.Split(res.Descr.OaiDC.Descr, "\n")
            for _, v := range descrLines {
                v = strings.Trim(v, " \n\t")
                if (v != "") {
                    fmt.Println()
                    scanner := bufio.NewScanner(strings.NewReader(v))
                    for scanner.Scan() {
                        fmt.Printf("   %s\n", scanner.Text())
                    }
                    fmt.Println()
                }
            }
        } else {
            fmt.Printf("%s\n", res.Spec)
        }
        return true
    })
}
