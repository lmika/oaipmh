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
    invertMatch         *bool
    urnOnly             *bool
    valueOnly           *bool
    downloadWorkers     *int

    matchNode           RecordSearcher
    hits                int
    misses              int
}

// Callbacks for the HarvesterObserver

func (sc *SearchCommand) OnRecord(recordResult *RecordResult) {
    hasRes, res, err := sc.matchNode.SearchRecord(recordResult)
    if (err != nil) {
        log.Printf("Record %s: Error: %s\n", recordResult.Identifier(), err.Error())
    }

    var matches bool = hasRes
    var showUrn bool = !*(sc.valueOnly)
    var showValue bool = !*(sc.urnOnly)
    if *(sc.invertMatch) {
        matches = !matches
        showValue = false       // Cannot show the value of unmatching result
    }

    // Display the results
    if (matches) {
        if showUrn && showValue {
            fmt.Printf("%s: %s\n", recordResult.Identifier(), res)
        } else if showUrn {
            fmt.Printf("%s\n", recordResult.Identifier())
        } else if showValue {
            fmt.Printf("%s\n", res)
        }
        sc.hits++
    } else {
        sc.misses++
    }
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
    if set == "" {
        set = sc.Ctx.Provider.Set
    } else if set == "*" {
        set = ""
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
        return &FileHarvester{
            Session:        sc.Ctx.Session,
            Filename:       *(sc.fromFile),
            FirstResult:    *(sc.firstResult),
            MaxResults:     *(sc.maxResults),
            Workers:        *(sc.downloadWorkers),
            Guard:          LiveRecordsPredicate,
        }
    } else if  (sc.listAndGet != nil) && *(sc.listAndGet) {
        return &ListAndGetRecordHarvester{
            Session:        sc.Ctx.Session,
            ListArgs:       la,
            FirstResult:    *(sc.firstResult),
            MaxResults:     *(sc.maxResults),
            Workers:        *(sc.downloadWorkers),
            HarvestGuard:   LiveRecordsHeaderPredicate,
            Guard:          LiveRecordsPredicate,
        }

    } else {
        return &ListRecordHarvester{
            Session:        sc.Ctx.Session,
            ListArgs:       la,
            FirstResult:    *(sc.firstResult),
            MaxResults:     *(sc.maxResults),
            Guard:          LiveRecordsPredicate,
        }
    }
}

// Startup flags
func (sc *SearchCommand) Flags(fs *flag.FlagSet) *flag.FlagSet {
    sc.setName = fs.String("s", "", "Select records from this set")
    sc.listAndGet = fs.Bool("L", false, "Use list and get instead of ListRecord")
    sc.beforeDate = fs.String("B", "", "Select records that were updated before date (YYYY-MM-DD)")
    sc.afterDate = fs.String("A", "", "Select records that were updated after date (YYYY-MM-DD)")
    sc.firstResult = fs.Int("f", 0, "Index of first record to retrieve")
    sc.fromFile = fs.String("F", "", "Read identifiers from a file")
    sc.maxResults = fs.Int("c", 100000, "Maximum number of records to retrieve")
    sc.invertMatch = fs.Bool("v", false, "Inverts the match.  Implies -h")
    sc.urnOnly = fs.Bool("l", false, "Only show the URN")
    sc.valueOnly = fs.Bool("h", false, "Only show the value")
    sc.downloadWorkers = fs.Int("W", 4, "Number of download workers running in parallel")

    return fs
}

// Runs the harvester
func (sc *SearchCommand) Run(args []string) {
    if (len(args) != 1) {
        fmt.Fprintf(os.Stderr, "Usage: cmdsearch <expr>\n")
        os.Exit(1)
    }

    // Attempt to parse the expression
    matchNode, err := ParseRecordMatchExpr(args[0])
    if (err != nil) {
        log.Fatal(err)
    }

    sc.matchNode = matchNode

    harvester := sc.makeHarvester()
    harvester.Harvest(sc)
}
