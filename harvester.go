package main

// A metadata harvester.  This is used for reading records from a remote OAI-PMH.
//

import (
    "fmt"
    "github.com/lmika-bom/oaipmh/mapreduce"
)


// Base interface for an harvester.
type Harvester interface {

    // Starts a harvesting task.
    Harvest(observer HarvesterObserver)
}


// Base interface for a recipient of harvested records.
type HarvesterObserver interface {
    // Called when a new record has been harvested.
    OnRecord(recordResult *RecordResult)

    // Called when an error is encountered.  The harvester may or may not continue.
    OnError(err error)

    // Called when the harvesting task has finished.
    OnCompleted(harvested int, skipped int, errors int)
}


// A predicate which matches headers.
type HeaderPredicate    func(hr *HeaderResult) bool

// A predicate which will select records.
type RecordPredicate    func(rr *RecordResult) bool


// A header predicate which will select all records
func AllRecordsHeaderPredicate(hr *HeaderResult) bool {
    return true
}

// A records predicate which will select all records
func AllRecordsPredicate(rr *RecordResult) bool {
    return true
}

// A predicate which will only select records that are not deleted
func LiveRecordsHeaderPredicate(rr *HeaderResult) bool {
    return !rr.Deleted
}

// A records predicate which will select only records that are live
func LiveRecordsPredicate(rr *RecordResult) bool {
    return !rr.Deleted
}

// --------------------------------------------------------------------------
// A few specialised observers that are used internally by the harvesters.


// An observer that simply counts the number of records and errors encountered.
// Has an optional predicate which can be used to determine if a record is selected
// or not.
type CountingObserver struct {
    Predicate       RecordPredicate

    Selected        int
    Skipped         int
    Errors          int
}

func (co *CountingObserver) OnRecord(rr *RecordResult) {
    if !co.Predicate(rr) {
        co.Skipped++
    } else {
        co.Selected++
    }
}

func (co *CountingObserver) OnError(err error) {
    co.Errors++
}

func (co *CountingObserver) OnCompleted(harvested int, skippedDeleted int, errors int) {
}


// A collection of observers.  When called, this will forward the result to the other
// observers.
type HarvesterObservers     []HarvesterObserver


func (os HarvesterObservers) OnRecord(recordResult *RecordResult) {
    for _, o := range os {
        o.OnRecord(recordResult)
    }
}

func (os HarvesterObservers) OnError(err error) {
    for _, o := range os {
        o.OnError(err)
    }
}

func (os HarvesterObservers) OnCompleted(harvested int, skippedDeleted int, errors int) {
    for _, o := range os {
        o.OnCompleted(harvested, skippedDeleted, errors)
    }
}


// --------------------------------------------------------------------------
// ListRecordHarvester
//      A harvester which uses the OAI-PMH "ListRecord" query.

type ListRecordHarvester struct {
    Session         *OaipmhSession
    ListArgs        ListIdentifierArgs
    FirstResult     int
    MaxResults      int

    Guard           RecordPredicate
}

// Starts the harvesting task.
func (rh *ListRecordHarvester) Harvest(observer HarvesterObserver) {
    pred := rh.Guard
    if pred == nil {
        pred = AllRecordsPredicate
    }


    var harvested, skipped, errors int = 0, 0, 0
    err := rh.Session.ListRecords(rh.ListArgs, rh.FirstResult, rh.MaxResults, func(rr *RecordResult) bool {
        if (pred(rr)) {
            observer.OnRecord(rr)
            harvested++
        } else {
            skipped++
        }
        return true
    })

    if (err != nil) {
        observer.OnError(err)
        errors++
    }

    observer.OnCompleted(harvested, skipped, errors)
}

// --------------------------------------------------------------------------
// ListAndGetRecordHarvester
//      A harvester which uses the OAI-PMH "ListIdentifier" and "ListRecord" query.
//      This harvester is a parallel based harvester as well.

type ListAndGetRecordHarvester struct {
    Session         *OaipmhSession
    ListArgs        ListIdentifierArgs
    FirstResult     int
    MaxResults      int
    Workers         int

    // A guard which will only queue records for harvesting with headers that match
    // this predicate.
    HarvestGuard    HeaderPredicate

    Guard           RecordPredicate
}

// Starts the harvesting task.
func (lgh *ListAndGetRecordHarvester) Harvest(observer HarvesterObserver) {
    pred := lgh.Guard
    if pred == nil {
        pred = AllRecordsPredicate
    }

    headPred := lgh.HarvestGuard
    if headPred == nil {
        headPred = AllRecordsHeaderPredicate
    }

    // The map reducer does not keep track of records and errors, so embed an observer
    // to do this.
    countingObserver := &CountingObserver{ Predicate: pred }
    observers := HarvesterObservers([]HarvesterObserver { observer, countingObserver })

    mr := newGetRecordMapReducer(lgh.Session, observers, lgh.Workers, pred)
    mr.Start()

    // Feed the data
    err := lgh.Session.ListIdentifiers(lgh.ListArgs, lgh.FirstResult, lgh.MaxResults, func(res *HeaderResult) bool {
        if (headPred(res)) {
            mr.Push(res.Identifier())
            return true
        } else {
            countingObserver.Skipped++
            return true
        }
    })
    mr.Close()

    if err != nil {
        observers.OnError(err)
    }

    // Send the completed signal
    observer.OnCompleted(countingObserver.Selected, countingObserver.Skipped, countingObserver.Errors)
}

// ----------------------------------------------------------------------
// FileHarvest
//      A harvester which will read records from a file.  This will harvest the
//      records in parallel.

type FileHarvester struct {
    Session             *OaipmhSession
    Filename            string
    FirstResult         int
    MaxResults          int
    Workers             int

    Guard               RecordPredicate
}

// Starts the harvesting task.
func (fh *FileHarvester) Harvest(observer HarvesterObserver) {

    pred := fh.Guard
    if pred == nil {
        pred = AllRecordsPredicate
    }

    // The map reducer does not keep track of records and errors, so embed an observer
    // to do this.
    countingObserver := &CountingObserver{ Predicate: pred }
    observers := HarvesterObservers([]HarvesterObserver { observer, countingObserver })

    mr := newGetRecordMapReducer(fh.Session, observers, fh.Workers, pred)
    mr.Start()

    // Feed the data
    err := LinesFromFile(fh.Filename, fh.FirstResult, fh.MaxResults, func(id string) bool {
        mr.Push(id)
        return true
    })
    mr.Close()

    if err != nil {
        observers.OnError(err)
    }

    // Send the completed signal
    observer.OnCompleted(countingObserver.Selected, countingObserver.Skipped, countingObserver.Errors)
}


// ------------------------------------------------------------------------


// Sets up a map/reducer with the following configuration.
//
//      Mapper:     URN -> downloaded record OR error
//      Reducer:    (record OR error)s -> calls to the observer
//
// The map/reduce expects URN to be pushed to the map queue.
//
func newGetRecordMapReducer(session *OaipmhSession, observer HarvesterObserver, downloadWorkers int, pred RecordPredicate) *mapreduce.SimpleMapReduce {
    return mapreduce.NewSimpleMapReduce(downloadWorkers, 100, downloadWorkers * 5).
            Map(func (id interface{}) interface{} {
                rec, err := session.GetRecord(id.(string))
                if (err == nil) {
                    return RecordToRecordResult(rec)
                } else {
                    return err
                }
            }).
            Reduce(func (recs chan interface{}) {
                // Retrieves either a *RecordResult or an error
                for rec := range recs {
                    switch r := rec.(type) {
                        case *RecordResult:
                            if (pred(r)) {
                                observer.OnRecord(r)
                            }
                        case error:
                            observer.OnError(r)
                        default:
                            panic(fmt.Sprintf("Expected either an recordResult or an error: got %v", rec))
                    }
                }
            })
}
