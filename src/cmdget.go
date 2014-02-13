package main


import (
    "fmt"
    "os"
    "flag"
    "log"
)


// ---------------------------------------------------------------------------------------------------
// Get command
//      Used to retrieve records.

type GetCommand struct {
    Ctx             *Context
    header          *bool
    wholeResponse   *bool
    separator       *string
    count           int
}

func (gc *GetCommand) Flags(fs *flag.FlagSet) *flag.FlagSet {
    gc.header = fs.Bool("H", false, "Display the header.")
    gc.wholeResponse = fs.Bool("R", false, "Display the entire OAI-PMH response.")
    gc.separator = fs.String("s", "====", "The record separator.")

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
        var file *os.File
        if (idExpr[1:] == "-") {
            file = os.Stdin
        } else {
            file, err := os.Open(idExpr[1:])
            if (file == nil) {
                panic(err)
            }
            defer file.Close()
        }

        var id string
        for num, _ := fmt.Fscanln(file, &id); num == 1; num, _ = fmt.Fscanln(file, &id) {
            callback(id)
        }
    } else {
        callback(idExpr)
    }
}

func (gc *GetCommand) displayRecord(id string) {
    if (gc.count >= 1) {
        fmt.Printf("%s\n", *(gc.separator))
    }

    if *(gc.wholeResponse) {
        payload, err := gc.Ctx.Session.GetRecordPayload(id)
        if (err == nil) {
            fmt.Println(payload)
        } else {
            log.Printf("Error: Cannot get record '%s', %s", id, err.Error())
        }
    } else {
        resp, err := gc.Ctx.Session.GetRecord(id)
        if (err == nil) {
            if *(gc.header) {
                for _, header := range resp.Header {
                    fmt.Printf("%s: %s\n", header[0], header[1])
                }
            } else {
                fmt.Print(resp.Content)
            }
        } else {
            log.Printf("Error: Cannot get record '%s', %s", id, err.Error())
        }
    }
    gc.count++
}
