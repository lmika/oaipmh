package main


import (
    "fmt"
    "os"
    "flag"
    "time"
)


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
    var set string

    // Get the set.  If '-s' is not provided and a provider is used, use the default set.
    // Otherwise, search all sets.  This implies that if using a provider, '-s ""' must be
    // used to search all sets.
    if lc.setName != nil {
        set = *(lc.setName)
    } else {
        set = lc.Ctx.Provider.Set
    }

    args := ListIdentifierArgs{
        Set: set,
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
        fmt.Printf("%-20s ", res.Sets[0])
        fmt.Printf("%-20s  ", res.Datestamp)
        fmt.Printf("%s\n", res.Identifier)
        return true
    })
}

func (lc *ListCommand) Flags(fs *flag.FlagSet) *flag.FlagSet {
    lc.setName = fs.String("s", "", "Select records from this set")
    lc.beforeDate = fs.String("B", "", "Select records that were updated before date (YYYY-MM-DD)")
    lc.afterDate = fs.String("A", "", "Select records that were updated after date (YYYY-MM-DD)")
    lc.flagDetailed = fs.Bool("l", false, "Use detailed listing format")
    lc.firstResult = fs.Int("f", 0, "Index of first record to retrieve")
    lc.maxResults = fs.Int("c", 100000, "Maximum number of records to retrieve")

    return fs
}

func (lc *ListCommand) Run(args []string) {
    if *(lc.flagDetailed) {
        lc.listIdentifiersInDetail()
    } else {
        lc.listIdentifiers()
    }
}
