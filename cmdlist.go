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
    listRecords     *bool
    showDeleted     *bool
    onlyShowDeleted *bool
}


type listingFn  func(listArgs ListIdentifierArgs, firstResult int, maxResults int, callback func(res *HeaderResult) bool) error


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
    set = *(lc.setName)
    if set == "" {
        set = lc.Ctx.Provider.Set
    } else if set == "*" {
        set = ""
    }

    args := ListIdentifierArgs{
        Set: set,
        From: parseDateString(*(lc.afterDate)),
        Until: parseDateString(*(lc.beforeDate)),
    }

    return args
}


// Returns the appropriate listing function.
func (lc *ListCommand) getListingFn() listingFn {
    if *(lc.listRecords) {
        return listingFn(lc.Ctx.Session.ListIdentifiersUsingListRecords)
    } else {
        return listingFn(lc.Ctx.Session.ListIdentifiers)
    }
}

// Returns true if the specific header should be shown.  This uses the options specified
// by the user.
func (lc *ListCommand) configuredToShowRecord(header *HeaderResult) bool {
    if *lc.showDeleted {
        return true
    } else if *lc.onlyShowDeleted {
        return header.Deleted
    } else {
        return !header.Deleted
    }
}


// List the identifiers from a provider
func (lc *ListCommand) listIdentifiers() {
    var deletedCount int = 0

    args := lc.genListIdentifierArgsFromCommandLine()
    listFn := lc.getListingFn()

    err := listFn(args, *(lc.firstResult), *(lc.maxResults), func(res *HeaderResult) bool {
        if lc.configuredToShowRecord(res) {
            fmt.Printf("%s\n", res.Identifier())
        }

        if (res.Deleted) {
            deletedCount += 1
        }

        return true
    })

    if (err == nil) {
        if (deletedCount > 0) {
            if *(lc.showDeleted) {
                // If onlyShowDeleted is active, the user expects the deleted records displayed so
                // don't bother showing anything.
                fmt.Fprintf(os.Stderr, "oaipmh: %d deleted record(s) displayed\n", deletedCount)
            } else if !*lc.onlyShowDeleted {
                fmt.Fprintf(os.Stderr, "oaipmh: %d deleted record(s) not displayed\n", deletedCount)
            }
        }
    } else {
        fmt.Fprintf(os.Stderr, "oaipmh: %s\n", err.Error())
    }


}

// List the identifiers in detail from a provider
func (lc *ListCommand) listIdentifiersInDetail() {
    var activeCount int = 0
    var deletedCount int = 0

    args := lc.genListIdentifierArgsFromCommandLine()
    listFn := lc.getListingFn()

    listFn(args, *(lc.firstResult), *(lc.maxResults), func(res *HeaderResult) bool {
        if (res.Deleted) {
            deletedCount++
        } else {
            activeCount++
        }

        if lc.configuredToShowRecord(res) {
            if (res.Deleted) {
                fmt.Printf("D ")
            } else {
                fmt.Printf(". ")
            }

            fmt.Printf("%-20s ", res.Header.SetSpec[0])
            fmt.Printf("%-20s  ", res.Header.DateStamp.String())
            fmt.Printf("%s\n", res.Identifier())
        }
        return true
    })

    fmt.Fprintf(os.Stderr, "%d records: %d active, %d deleted\n", activeCount + deletedCount, activeCount, deletedCount)
}

func (lc *ListCommand) Flags(fs *flag.FlagSet) *flag.FlagSet {
    lc.setName = fs.String("s", "", "Select records from this set")
    lc.beforeDate = fs.String("B", "", "Select records that were updated before date (YYYY-MM-DD)")
    lc.afterDate = fs.String("A", "", "Select records that were updated after date (YYYY-MM-DD)")
    lc.flagDetailed = fs.Bool("l", false, "Use detailed listing format")
    lc.showDeleted = fs.Bool("d", false, "Show deleted records, along with active ones")
    lc.onlyShowDeleted = fs.Bool("D", false, "Only show deleted records")
    lc.firstResult = fs.Int("f", 0, "Index of first record to retrieve")
    lc.maxResults = fs.Int("c", 100000, "Maximum number of records to retrieve")
    lc.listRecords = fs.Bool("R", false, "Use ListRecord instead of ListIdentifier")

    return fs
}

func (lc *ListCommand) Run(args []string) {
    if *(lc.flagDetailed) {
        lc.listIdentifiersInDetail()
    } else {
        lc.listIdentifiers()
    }
}
