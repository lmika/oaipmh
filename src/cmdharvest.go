package main


import (
    "fmt"
    "os"
    "flag"
    "time"
)


// ---------------------------------------------------------------------------------------------------
// Harvest commands
//      Extract the records from a provider and store them in a directory.

type HarvestCommand struct {
    Ctx                 *Context
    dryRun              *bool
    listAndGet          *bool
    setName             *string
    beforeDate          *string
    afterDate           *string
    fromFile            *string
    firstResult         *int
    maxResults          *int
    maxDirSize          *int
    downloadWorkers     *int

    dirPrefix           string
    recordCount         int
    deletedCount        int
}

// Get list identifier arguments
func (lc *HarvestCommand) genListIdentifierArgsFromCommandLine() ListIdentifierArgs {
    set := *(lc.setName)
    if (set == "") {
        set = lc.Ctx.Provider.Set
    }

    args := ListIdentifierArgs{
        Set: set,
        From: parseDateString(*(lc.afterDate)),
        Until: parseDateString(*(lc.beforeDate)),
    }

    return args
}

// Saves the record
func (lc *HarvestCommand) saveRecord(dirId int, res *RecordResult) {
    id := res.Identifier()
    dir := fmt.Sprintf("%s/%02d", lc.dirPrefix, dirId)
    outFile := fmt.Sprintf("%s/%s.xml", dir, id)

    os.MkdirAll(dir, 0755)

    file, err := os.Create(outFile)
    if err != nil {
        panic(err)
    }
    defer file.Close()

    file.WriteString("<?xml version=\"1.0\"?>\n")
    file.WriteString(res.Content)
}

// Handle the record harvested
func (lc *HarvestCommand) withRecord(res *RecordResult) bool {
    if (! res.Deleted) {
        lc.recordCount++
        dirId := (lc.recordCount / *(lc.maxDirSize)) + 1

        fmt.Printf("%8d  %s\n", lc.recordCount, res.Identifier())
        if (! *(lc.dryRun)) {
            lc.saveRecord(dirId, res)
        }
    } else {
        lc.deletedCount++
    }
    return true
}

// Setup a map reduce parallel worker for downloading records from a source.  The mapping
// function is expected to be given URNs.
func (lc *HarvestCommand) setupParallelHarvester() *SimpleMapReduce {
    return NewSimpleMapReduce(*(lc.downloadWorkers), 100, *(lc.downloadWorkers) * 5).
            Map(func (id interface{}) interface{} {
                return lc.Ctx.Session.GetRecord(id.(string))
            }).
            Reduce(func (recs chan interface{}) {
                for rec, hasMore := <-recs ; hasMore ; rec, hasMore = <-recs {
                    lc.withRecord(rec.(*RecordResult))
                }
            }).
            Start()
}

// List the identifiers from a provider
func (lc *HarvestCommand) harvest() {
    args := lc.genListIdentifierArgsFromCommandLine()

    if *(lc.fromFile) != "" {
        // Setup a map-reduce queue for fetching responses in parallel
        mr := lc.setupParallelHarvester()

        // Push records from a file
        LinesFromFile(*(lc.fromFile), *(lc.firstResult), *(lc.maxResults), func(id string) bool {
            mr.Push(id)
            return true
        })
        mr.Close()

    } else if *(lc.listAndGet) {
        // Get the list and pass it to the getters in parallel
        mr := lc.setupParallelHarvester()

        lc.Ctx.Session.ListIdentifiers(args, *(lc.firstResult), *(lc.maxResults), func(res ListIdentifierResult) bool {
            if (! res.Deleted) {
                mr.Push(res.Identifier)
                return true
            } else {
                lc.deletedCount++
                return true
            }
        })

        mr.Close()
    } else {
        lc.Ctx.Session.ListRecords(args, *(lc.firstResult), *(lc.maxResults), lc.withRecord)
    }

    if (lc.deletedCount > 0) {
        fmt.Fprintf(os.Stderr, "oaipmh: %d deleted record(s) not harvested.\n", lc.deletedCount)
    }

}

func (lc *HarvestCommand) Flags(fs *flag.FlagSet) *flag.FlagSet {
    lc.dryRun = fs.Bool("n", false, "Dry run.  Do not save the results to a file.")
    lc.listAndGet = fs.Bool("L", false, "Use list and get instead of ListRecord.  Slow.")
    lc.setName = fs.String("s", "", "The set to retrieve")
    lc.beforeDate = fs.String("B", "", "List metadata records that have been updated before this date (YYYY-MM-DD).")
    lc.afterDate = fs.String("A", "", "List metadata records that have been updated after this date (YYYY-MM-DD).")
    lc.firstResult = fs.Int("f", 0, "The first result to return.")
    lc.fromFile = fs.String("F", "", "Read identifiers from a file.")
    lc.maxResults = fs.Int("c", 100000, "Maximum number of results to return.")
    lc.maxDirSize = fs.Int("D", 10000, "Maximum number of files to store in each directory.")
    lc.downloadWorkers = fs.Int("W", 4, "Number of download workers running in parallel.")

    return fs
}

func (lc *HarvestCommand) Run(args []string) {
    lc.dirPrefix = time.Now().Format("20060102T150405")
    lc.harvest()
}
