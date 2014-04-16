// Interface for an OAI-PMH repository.
//

package oaipmh

import (
    "time"
)


// The minimum time to return records from if not specified.
var MinTime time.Time = time.Date(1900, 01, 01, 01, 01, 01, 01, time.UTC)


// Interface for an OAI-PMH repository.
type Repository interface {

    // Returns a slice of sets managed by this repository.
    Sets() ([]Set, error)

    // Returns a list of formats managed by this repository.
    Formats() []Format

    // Returns a record cursor which can be used for moving through the list of records over the 
    // repository.
    //      set     The set to return or "" if not specified.
    //      from    The time to start listing records from, or MIN_TIME if not specified.
    //      to      The time to end listing records from, or time.Now() if not specified.
    // The returned cursor is to be positioned at the first record (i.e. calling Record() without
    // calling Next() should return the first record).
    ListRecords(set string, from time.Time, to time.Time) (RecordCursor, error)

    // Returns a single record.
    Record(id string) (*Record, error)
}


// Interface for a record cursor.
type RecordCursor interface {
    // Indicates if the cursor is pointing to a record
    HasRecord() bool

    // Goes to the next record.  If the next record exists, returns true.  Otherwise, returns false.
    Next() bool

    // Moves the cursor to a particular position.  If the position is valid, returns true.
    SetPos(pos int) bool

    // Returns the current position of the cursor.
    Pos() int

    // Returns the current record, or nil if the cursor is at an invalid position.
    Record() *Record
}

// Returns the next N records from a cursor.  Returns true if there are more records to return.
func NextNRecords(cursor RecordCursor, n int) (records []*Record, hasmore bool) {
    if (n == 0) {
        records = make([]*Record, 0)
        hasmore = true
        return
    }

    records = make([]*Record, 0, n)
    for (n > 0) && (cursor.HasRecord()) {
        records = append(records, cursor.Record())
        cursor.Next()
        n--
    }

    hasmore = cursor.HasRecord()
    return
}


// Metadata formats
type Format struct {
    Prefix      string          `xml:"metadataPrefix"`
    Schema      string          `xml:"schema"`
    Namespace   string          `xml:"metadataNamespace"`
}

// Metadata sets
type Set struct {
    Spec        string
    Name        string
    Descr       string
}

// Metadata record
type Record struct {
    ID          string
    Date        time.Time
    Set         string

    // Function to call to the the content of the record.
    Content     func() (string, error)
}
