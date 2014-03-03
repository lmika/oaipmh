package main


import (
    "fmt"
    "os"
    "os/exec"
    "path"
    "flag"
    "time"
    "log"
)


// ---------------------------------------------------------------------------------------------------
// Harvest commands
//      Extract the records from a provider and store them in a directory.

type HarvestCommand struct {
    Ctx                 *Context
    dryRun              *bool
    listAndGet          *bool
    compressDirs        *bool
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
    errorCount          int
    lastDirId           int
}

// Get list identifier arguments
func (lc *HarvestCommand) genListIdentifierArgsFromCommandLine() ListIdentifierArgs {
    var set string

    if (lc.setName == nil) {
        set = lc.Ctx.Provider.Set
    } else {
        set = *(lc.setName)
    }

    args := ListIdentifierArgs{
        Set: set,
        From: parseDateString(*(lc.afterDate)),
        Until: parseDateString(*(lc.beforeDate)),
    }

    return args
}

// Returns the name of directory given the directory ID
func (lc *HarvestCommand) dirName(dirId int) string {
    return fmt.Sprintf("%s/%02d", lc.dirPrefix, dirId)
}

// Saves the record
func (lc *HarvestCommand) saveRecord(dirId int, res *RecordResult) {
    id := res.Identifier()
    dir := lc.dirName(dirId)
    outFile := fmt.Sprintf("%s/%s.xml", dir, id)

    os.MkdirAll(dir, 0755)

    file, err := os.Create(outFile)
    if err != nil {
        panic(err)
    }
    defer file.Close()

    file.WriteString(res.Content)
}

// Close the current directory before creating and writing to a new one
func (lc *HarvestCommand) closeDir(dirId int) {
    // Do nothing if this is a dry run
    if *(lc.dryRun) {
        return
    }

    dir := lc.dirName(dirId)
    if *(lc.compressDirs) {
        base := path.Base(dir)
        parent := path.Dir(dir)

        if (lc.Ctx.Debug) {
            log.Printf("Compressing %s -> %s", base, dir + ".zip")
        }

        cmd := exec.Command("zip", "-m", "-r", base + ".zip", base)
        cmd.Dir = parent
        err := cmd.Start()
        if (err != nil) {
            fmt.Fprintf(os.Stderr, "Cannot compress '%s'\n", dir)
        }
    }
}

// Handle the record harvested
func (lc *HarvestCommand) withRecord(res *RecordResult) bool {
    if (! res.Deleted) {
        lc.recordCount++
        dirId := (lc.recordCount / *(lc.maxDirSize)) + 1
        if (dirId != lc.lastDirId) {
            lc.closeDir(lc.lastDirId)
            lc.lastDirId = dirId
        }

        if (lc.Ctx.Debug) {
            log.Printf("%8d  %s\n", lc.recordCount, res.Identifier())
        }
        if ((lc.recordCount % 1000) == 0) {
            log.Printf("Harvested %d records\n", lc.recordCount)
        }

        if (! *(lc.dryRun)) {
            lc.saveRecord(dirId, res)
        }
    } else {
        lc.deletedCount++
    }
    return true
}

// Handles an error returned
func (lc *HarvestCommand) withError(err error) {
    log.Printf("ERROR: %s\n", err)
    lc.errorCount++
}

// Setup a map reduce parallel worker for downloading records from a source.  The mapping
// function is expected to be given URNs.
func (lc *HarvestCommand) setupParallelHarvester() *SimpleMapReduce {
    return NewSimpleMapReduce(*(lc.downloadWorkers), 100, *(lc.downloadWorkers) * 5).
            Map(func (id interface{}) interface{} {
                rec, err := lc.Ctx.Session.GetRecord(id.(string))
                if (err == nil) {
                    return rec
                } else {
                    return err
                }
            }).
            Reduce(func (recs chan interface{}) {
                // Retrieves either a *RecordResult or an error
                for rec, hasMore := <-recs ; hasMore ; rec, hasMore = <-recs {
                    switch r := rec.(type) {
                        case *RecordResult:
                            lc.withRecord(r)
                        case error:
                            lc.withError(r)
                        default:
                            panic("Expected either an recordResult or an error")
                    }
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
    lc.dryRun = fs.Bool("n", false, "Dry run.  Do not save results.")
    lc.listAndGet = fs.Bool("L", false, "Use list and get instead of ListRecord")
    lc.setName = fs.String("s", "", "Select records from this set")
    lc.beforeDate = fs.String("B", "", "Select records that were updated before date (YYYY-MM-DD)")
    lc.afterDate = fs.String("A", "", "Select records that were updated after date (YYYY-MM-DD)")
    lc.firstResult = fs.Int("f", 0, "Index of first record to retrieve")
    lc.fromFile = fs.String("F", "", "Read identifiers from a file")
    lc.maxResults = fs.Int("c", 100000, "Maximum number of records to retrieve")
    lc.maxDirSize = fs.Int("D", 10000, "Maximum number of files to store in each directory")
    lc.compressDirs = fs.Bool("C", false, "Compress directories once they are full")
    lc.downloadWorkers = fs.Int("W", 4, "Number of download workers running in parallel")

    return fs
}

func (lc *HarvestCommand) Run(args []string) {
    lc.lastDirId = 1
    lc.dirPrefix = time.Now().Format("20060102T150405")
    lc.harvest()
    lc.closeDir(lc.lastDirId)

    log.Printf("Finished: %d records harvested, %d deleted records, %d errors", lc.recordCount, lc.deletedCount, lc.errorCount)
}
