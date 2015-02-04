package main


import (
    "fmt"
    "os"
    "bufio"
    "flag"
    "log"
    "strings"
    "encoding/xml"
)


// ---------------------------------------------------------------------------------------------------
// Get command
//      Used to retrieve records.

type GetCommand struct {
    Ctx             *Context
    header          *bool
    test            *bool
    separator       *string
    count           int
}

func (gc *GetCommand) Flags(fs *flag.FlagSet) *flag.FlagSet {
    gc.header = fs.Bool("H", false, "Display record header")
    gc.test = fs.Bool("t", false, "Test that the record can be retrieved")
    gc.separator = fs.String("s", "====", "Record separator")

    return fs
}

func (gc *GetCommand) Run(args []string) {
    for _, id := range args {
        gc.eachId(id, func(urn string) {
            gc.displayRecord(urn)
        })
    }
}

// Interprets an id and calls the callback with each interpred ID.
func (gc *GetCommand) eachId(idExpr string, callback func(string)) {
    if (idExpr[0] == '@') {
        var err error
        var file *os.File

        if (idExpr[1:] == "-") {
            file = os.Stdin
        } else {
            file, err = os.Open(idExpr[1:])
            if (err != nil) {
                panic(err)
            }
            defer file.Close()
        }

        lineReader := bufio.NewReader(file)
        for id, err := lineReader.ReadString('\n') ; err == nil ; id, err = lineReader.ReadString('\n') {
            id = strings.TrimSpace(id)
            callback(id)
        }
        if err != nil {
            panic(err)
        }
    } else {
        callback(idExpr)
    }
}

func (gc *GetCommand) displayRecord(id string) {
    if (gc.count >= 1) && (! *(gc.test)) {
        fmt.Printf("%s\n", *(gc.separator))
    }

    resp, err := gc.Ctx.Session.GetRecord(id)
    if (err == nil) {
        if *(gc.test) {
            fmt.Printf("+ %s\n", resp.Header.Identifier)
        } else if *(gc.header) {
            fmt.Printf("Id:\t%s\n", resp.Header.Identifier)
            fmt.Printf("Date:\t%s\n", resp.Header.DateStamp.String())
            fmt.Printf("Sets:\t%s\n", strings.Join(resp.Header.SetSpec, ", "))
        } else {
            fmt.Print(xml.Header)
            fmt.Println(strings.TrimSpace(resp.Content.Xml))
        }
    } else {
        if *(gc.test) {
            fmt.Printf("- %s\n", id)
        } else {
            log.Printf("Error: Cannot get record '%s', %s", id, err.Error())
        }
    }
    gc.count++
}
