package main


import (
    "fmt"
    "os"
    "flag"
    "time"
)


// Const
const DateFormat string = "2006-01-02"



type Context struct {
    Session         *OaipmhSession
}

// ---------------------------------------------------------------------------------------------------
// Sets commands

type SetsCommand struct {
    Ctx             *Context
}

func (lc *SetsCommand) Flags(fs *flag.FlagSet) *flag.FlagSet {
    return fs
}

func (lc *SetsCommand) Run(args []string) {
    lc.Ctx.Session.ListSets(0, -1, func(res ListSetResult) bool {
        fmt.Printf("%s\n", res.Spec)
        return true
    })
}

// ---------------------------------------------------------------------------------------------------
// List commands

type ListCommand struct {
    Ctx             *Context
    setName         *string
    beforeDate      *string
    afterDate       *string
    flagDetailed    *bool
    firstResult     *int
    maxResults      *int
}


// Attempt to parse a date string and return it as a heap allocated time.Time.
// If the string is empty, returns nil.  If there was an error parsing the string,
// the program dies
func parseDateString(dateString string) *time.Time {
    if (dateString != "") {
        parsedTime, err := time.ParseInLocation(DateFormat, dateString, time.Local)
        if (err != nil) {
            die("Invalid date: " + err.Error())
        }

        heapAllocTime := new(time.Time)
        *heapAllocTime = parsedTime

        return heapAllocTime
    } else {
        return nil
    }

}



// Get list identifier arguments
func (lc *ListCommand) genListIdentifierArgsFromCommandLine() ListIdentifierArgs {
    args := ListIdentifierArgs{
        Set: *(lc.setName),
        From: parseDateString(*(lc.afterDate)),
        Until: parseDateString(*(lc.beforeDate)),
    }

    return args
}

// List the identifiers from a provider
func (lc *ListCommand) listIdentifiers() {
    var deletedCount int = 0

    args := lc.genListIdentifierArgsFromCommandLine()

    lc.Ctx.Session.ListIdentifiers(args, *(lc.firstResult), *(lc.maxResults), func(res ListIdentifierResult) bool {
        if (res.Deleted) {
            deletedCount += 1
        } else {
            fmt.Printf("%s\n", res.Identifier)
        }
        return true
    })

    if (deletedCount > 0) {
        fmt.Fprintf(os.Stderr, "oaipmh: %d deleted record(s) not displayed.\n", deletedCount)
    }

}

// List the identifiers in detail from a provider
func (lc *ListCommand) listIdentifiersInDetail() {
    args := lc.genListIdentifierArgsFromCommandLine()

    lc.Ctx.Session.ListIdentifiers(args, *(lc.firstResult), *(lc.maxResults), func(res ListIdentifierResult) bool {
        if (res.Deleted) {
            fmt.Printf("D ")
        } else {
            fmt.Printf(". ")
        }
        fmt.Printf("%-20s ", res.Set)
        fmt.Printf("%-20s  ", res.Datestamp)
        fmt.Printf("%s\n", res.Identifier)
        return true
    })
}

func (lc *ListCommand) Flags(fs *flag.FlagSet) *flag.FlagSet {
    lc.setName = fs.String("s", "", "The set to retrieve")
    lc.beforeDate = fs.String("B", "", "List metadata records that have been updated since this date (YYYY-MM-DD).")
    lc.afterDate = fs.String("A", "", "List metadata records that have been updated after this date (YYYY-MM-DD).")
    lc.flagDetailed = fs.Bool("l", false, "List metadata in detail.")
    lc.firstResult = fs.Int("f", 0, "The first result to return.")
    lc.maxResults = fs.Int("c", 10000, "Maximum number of results to return.")

    return fs
}

func (lc *ListCommand) Run(args []string) {
    if *(lc.flagDetailed) {
        lc.listIdentifiersInDetail()
    } else {
        lc.listIdentifiers()
    }
}

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

    if *(gc.header) {
        var headers [][]string = gc.Ctx.Session.GetRecordHeader(id)
        for _, header := range headers {
            fmt.Printf("%s: %s\n", header[0], header[1])
        }
    } else if *(gc.wholeResponse) {
        fmt.Print(gc.Ctx.Session.GetRecord(id))
    } else {
        fmt.Println(gc.Ctx.Session.GetRecordPayload(id))
    }
    gc.count++
}
