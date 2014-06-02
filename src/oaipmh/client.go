// Client package for an OAI-PMH
//

package oaipmh

import (
    "log"
    "encoding/xml"
    "fmt"
    "net/http"
    "net/url"
    "time"
)

// An error indicating that there are no more results in an iterator
type ENoMore struct{}

func (e ENoMore) Error() string {
    return "No more results"
}



// An OAI-PMH error
type EOaipmhError struct {
    Code        string
    Message     string
}

func (e EOaipmhError) Error() string {
    return fmt.Sprintf("OAI-PMH Error (%s): %s", e.Code, e.Message)
}


// Arguments for ListIdentifier and ListRecords
type ListArgs struct {
    Prefix          string
    Set             string
    From            *time.Time
    Until           *time.Time
}


// An OAI-PMH client
type Client struct {
    // If true, will display debug information
    Debug           bool

    url             *url.URL
}

// Creates a new client to a particular provider.  Returns either the client or an
// error if the URL is invalid
func NewClient(providerUrl string) (*Client, error) {
    u, err := url.ParseRequestURI(providerUrl)
    if (err != nil) {
        return nil, err
    }

    return &Client{false, u}, nil
}

// Fetches an OAI-PMH request and stores it within the provider response variable.  Returns
// an error if there was an error.
func (c *Client) Fetch(verb string, vals url.Values, res *OaipmhResponse) error {
    vals.Set("verb", verb)

    if (c.Debug) {
        log.Printf(">> POST %s\n", c.url.String() + "?" + vals.Encode())
    }

    // Post the form
    resp, err := http.PostForm(c.url.String(), vals)
    if err != nil {
        return err
    }
    defer resp.Body.Close()

    // Expect a 200 response
    if resp.StatusCode != 200 {
        return fmt.Errorf("HTTP error: %d\n", resp.Status)
    }

    // Marshal the response into the provided 'res'
    dec := xml.NewDecoder(resp.Body)
    err = dec.Decode(res)
    if err != nil {
        return err
    }

    // If there's an OAI-PMH error, return that as a normal error.
    if res.Error != nil {
        return EOaipmhError{res.Error.Code, res.Error.Message}
    }

    return nil
}

// Returns the list of sets
func (c *Client) ListSets() ([]OaipmhSet, error) {
    res := &OaipmhResponse{}
    err := c.Fetch("ListSets", url.Values{}, res)
    if (err != nil) {
        return nil, err
    }

    return res.ListSets.Sets, nil
}

// Returns a record
func (c *Client) GetRecord(prefix string, id string) (*OaipmhRecord, error) {
    res := &OaipmhResponse{}
    err := c.Fetch("GetRecord", url.Values{
        "metadataPrefix": { prefix },
        "identifier": { id },
    }, res)
    if (err != nil) {
        return nil, err
    }

    return &(res.GetRecord.Record), nil
}

// Returns a list of identifiers
func (c *Client) ListIdentifiers(listArgs ListArgs) (*ListIdentifierIterator, error) {
    vals := url.Values{
        "metadataPrefix": {listArgs.Prefix},
    }
    if (listArgs.From != nil) {
        vals.Add("from", listArgs.From.UTC().Format(time.RFC3339))
    }
    if (listArgs.Until != nil) {
        vals.Add("until", listArgs.Until.UTC().Format(time.RFC3339))
    }
    if (listArgs.Set != "") {
        vals.Set("set", listArgs.Set)
    }

    // Get the initial set
    li := &ListIdentifierIterator{ client: c }
    err := li.fetch(vals)
    if (err != nil) {
        return nil, err
    }

    return li, nil
}

// Returns a list of records
func (c *Client) ListRecords(listArgs ListArgs) (*ListRecordsIterator, error) {
    vals := url.Values{
        "metadataPrefix": {listArgs.Prefix},
    }
    if (listArgs.From != nil) {
        vals.Add("from", listArgs.From.UTC().Format(time.RFC3339))
    }
    if (listArgs.Until != nil) {
        vals.Add("until", listArgs.Until.UTC().Format(time.RFC3339))
    }
    if (listArgs.Set != "") {
        vals.Set("set", listArgs.Set)
    }

    // Get the initial set
    lr := &ListRecordsIterator{ client: c }
    err := lr.fetch(vals)
    if (err != nil) {
        return nil, err
    }

    return lr, nil
}

// -----------------------------------------------------------------------------------------
// RecordIterator
//      An iterator which will iterate results over a ListIdentifier and RecordIdentifier
//      response.

type RecordIterator interface {

    // Returns the next record.  If a record is present, this returns nil.  If there are
    // no more records, an ENoMore is returned.  Otherwise, this returns an error.
    // This must be called first before calling Header or Record.
    Next()          error

    // Returns the current header.  If Next() returns nil, this is guaranteed to be set.
    Header()        (*OaipmhHeader, error)

    // Returns the current record.  Either returns the record, or an error if there was a
    // problem retrieving the record.  The record may be retrieved on demand if necessary.
    Record()        (*OaipmhRecord, error)
}

// -----------------------------------------------------------------------------------------
// ListIdentifiers iterator

// An iterator for a list identifiers response
type ListIdentifierIterator struct {
    client          *Client
    resToken        string
    pos             int             // Position will be 1 ahead of current header
    headers         []OaipmhHeader
}

// Returns the next header, if one is present.  If no more headers are present, the second
// return value will be a ENoMore result.  Otherwise, the error will be something else.
func (li *ListIdentifierIterator) Next() error {
    if (li.pos >= len(li.headers)) {
        if (li.resToken == "") {
            return ENoMore{}
        }

        err := li.fetchNext()
        if (err != nil) {
            return err
        } else {
            return li.Next()
        }
    } else {
        li.pos++
        return nil
    }
}

// Returns the current header.  If Next() returns nil, this is guaranteed to be set.
func (li *ListIdentifierIterator) Header() (*OaipmhHeader, error) {
    if (li.pos > 0) {
        return &(li.headers[li.pos - 1]), nil
    } else {
        return nil, fmt.Errorf("Next() was not called first")
    }
}

// Returns the current record.  Either returns the record, or an error if there was a
// problem retrieving the record.  The record may be retrieved on demand if necessary.
func (li *ListIdentifierIterator) Record() (*OaipmhRecord, error) {
    return nil, fmt.Errorf("Records are not fetched")
}

// Loads the iterator with new values
func (li *ListIdentifierIterator) fetchNext() error {
    return li.fetch(url.Values{ "resumptionToken": {li.resToken} })
}

// Fetch the next set of headers
func (li *ListIdentifierIterator) fetch(val url.Values) error {
    res := &OaipmhResponse{}

    err := li.client.Fetch("ListIdentifiers", val, res)

    if (err != nil) {
        return err
    }

    liRes := res.ListIdentifiers

    // Set the new state
    li.pos = 0
    li.headers = liRes.Headers
    li.resToken = liRes.ResumptionToken

    return nil
}

// -----------------------------------------------------------------------------------------
// ListRecords iterator

// An iterator for a list records response
type ListRecordsIterator struct {
    client          *Client
    resToken        string
    pos             int
    records         []OaipmhRecord
}

// Returns the next record, if one is present.  If no more records are present, the second
// return value will be a ENoMore result.  Otherwise, the error will be something else.
func (lr *ListRecordsIterator) Next() error {
    if (lr.pos >= len(lr.records)) {
        if (lr.resToken == "") {
            return ENoMore{}
        }

        err := lr.fetchNext()
        if (err != nil) {
            return err
        } else {
            return lr.Next()
        }
    } else {
        lr.pos++
        return nil
    }
}

// Returns the current header.  If Next() returns nil, this is guaranteed to be set.
func (lr *ListRecordsIterator) Header() (*OaipmhHeader, error) {
    if (lr.pos > 0) {
        return &(lr.records[lr.pos - 1].Header), nil
    } else {
        return nil, fmt.Errorf("Next() was not called first")
    }
}

// Returns the current record.  Either returns the record, or an error if there was a
// problem retrieving the record.  The record may be retrieved on demand if necessary.
func (lr *ListRecordsIterator) Record() (*OaipmhRecord, error) {
    if (lr.pos > 0) {
        return &(lr.records[lr.pos - 1]), nil
    } else {
        return nil, fmt.Errorf("Next() was not called first")
    }
}

// Loads the iterator with new values
func (lr *ListRecordsIterator) fetchNext() error {
    return lr.fetch(url.Values{ "resumptionToken": {lr.resToken} })
}

// Fetch the next set of headers
func (lr *ListRecordsIterator) fetch(val url.Values) error {
    res := &OaipmhResponse{}

    err := lr.client.Fetch("ListRecords", val, res)

    if (err != nil) {
        return err
    }

    lrRes := res.ListRecords

    // Set the new state
    lr.pos = 0
    lr.records = lrRes.Records
    lr.resToken = lrRes.ResumptionToken

    return nil
}
