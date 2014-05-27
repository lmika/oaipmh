package main

// A metadata harvester.  This is used for reading records from a remote OAI-PMH.
//


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
    OnCompleted(harvested int, skippedDeleted int, errors int)
}


// --------------------------------------------------------------------------
// ListRecordHarvester
//      A harvester which uses the OAI-PMH "ListRecord" query.

type ListRecordHarvester struct {
    Session         *OaipmhSession
    ListArgs        ListIdentifierArgs
    FirstResult     int
    MaxResults      int
}

// Starts the harvesting task.
func (rh *ListRecordHarvester) Harvest(observer HarvesterObserver) {
    var harvested, skipped, errors int = 0, 0, 0
    err := rh.Session.ListRecords(rh.ListArgs, rh.FirstResult, rh.MaxResults, func(rr *RecordResult) bool {
        if (! rr.Deleted) {
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

