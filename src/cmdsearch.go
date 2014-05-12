package main

import (
    "log"
    "flag"
    "os"
    "fmt"
)

// --------------------------------------------------------------------------------
// Search command
//      Searches through the metadata from an endpoint for metadata records which
//      match an XPath expression.
//

type SearchCommand struct {
    Ctx                 *Context
    listAndGet          *bool
    setName             *string
    beforeDate          *string
    afterDate           *string
    fromFile            *string
    firstResult         *int
    maxResults          *int

    xpath               string
    hits                int
    misses              int
}

// Callbacks for the HarvesterObserver

func (sc *SearchCommand) OnRecord(recordResult *RecordResult) {
    panic("Not supported yet")
//    res, err := recordResult.RunXPath(sc.xpath)
//    if (err != nil) {
//        log.Printf("Record %s: XPath Error: %s\n", recordResult.Identifier(), err.Error())
//    } else if (res) {
 //       fmt.Println(recordResult.Identifier())
 //       sc.hits++
//    } else {
//        sc.misses++
//    }
}

func (sc *SearchCommand) OnError(err error) {
    log.Printf("Harvesting Error: %s\n", err.Error())
}

func (sc *SearchCommand) OnCompleted(harvested int, skipped int, errors int) {
    log.Printf("Search Complete: hits = %d, misses = %d, skips = %d, errors = %d\n", sc.hits, sc.misses, skipped, errors)
}

// Get list identifier arguments
func (sc *SearchCommand) genListIdentifierArgsFromCommandLine() ListIdentifierArgs {
    var set string

    set = *(sc.setName)
    if set == "\x00" {
        set = sc.Ctx.Provider.Set
    }

    args := ListIdentifierArgs{
        Set: set,
        From: parseDateString(*(sc.afterDate)),
        Until: parseDateString(*(sc.beforeDate)),
    }

    return args
}

// Build the harvester based on the configuration
func (sc *SearchCommand) makeHarvester() Harvester {
    la := sc.genListIdentifierArgsFromCommandLine()

    if *(sc.fromFile) != "" {
        panic("From file is not yet supported")
    } else if *(sc.listAndGet) {
        panic("ListAndGet is not yet supported")
    } else {
        return &ListRecordHarvester{
            Session: sc.Ctx.Session,
            ListArgs: la,
            FirstResult: *(sc.firstResult),
            MaxResults: *(sc.maxResults),
        }
    }
}

// Startup flags
func (sc *SearchCommand) Flags(fs *flag.FlagSet) *flag.FlagSet {
    sc.setName = fs.String("s", "\x00", "Select records from this set")
    sc.listAndGet = fs.Bool("L", false, "Use list and get instead of ListRecord")
    sc.beforeDate = fs.String("B", "", "Select records that were updated before date (YYYY-MM-DD)")
    sc.afterDate = fs.String("A", "", "Select records that were updated after date (YYYY-MM-DD)")
    sc.firstResult = fs.Int("f", 0, "Index of first record to retrieve")
    sc.fromFile = fs.String("F", "", "Read identifiers from a file")
    sc.maxResults = fs.Int("c", 100000, "Maximum number of records to retrieve")

    return fs
}

// Runs the harvester
func (sc *SearchCommand) Run(args []string) {
    if (len(args) != 1) {
        fmt.Fprintf(os.Stderr, "Usage: cmdsearch <xpath>\n")
    }

    sc.xpath = args[0]

    harvester := sc.makeHarvester()
    harvester.Harvest(sc)
}
