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


// Arguments for ListIdentifier and ListRecords
type ListArgs struct {
    Prefix          string
    Set             string
    From            *time.Time
    Until           *time.Time
}


// An OAI-PMH client
type Client struct {
    url             *url.URL
}

// Creates a new client to a particular provider.  Returns either the client or an
// error if the URL is invalid
func NewClient(providerUrl string) (*Client, error) {
    u, err := url.ParseRequestURI(providerUrl)
    if (err != nil) {
        return nil, err
    }

    return &Client{u}, nil
}

// Fetches an OAI-PMH request and stores it within the provider response variable.  Returns
// an error if there was an error.
func (c *Client) Fetch(verb string, vals url.Values, res interface{}) error {
    vals.Set("verb", verb)

    log.Printf(">> POST %s\n", c.url.String() + "?" + vals.Encode())

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
    return dec.Decode(res)
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
// ListIdentifiers iterator

// An iterator for a list identifiers response
type ListIdentifierIterator struct {
    client          *Client
    resToken        string
    pos             int
    headers         []OaipmhHeader
}

// Returns the next header, if one is present.  If no more headers are present, the second
// return value will be a ENoMore result.  Otherwise, the error will be something else.
func (li *ListIdentifierIterator) Next() (OaipmhHeader, error) {
    if (li.pos >= len(li.headers)) {
        if (li.resToken == "") {
            return OaipmhHeader{}, ENoMore{}
        }

        err := li.fetchNext()
        if (err != nil) {
            return OaipmhHeader{}, err
        } else {
            return li.Next()
        }
    } else {
        li.pos++
        return li.headers[li.pos - 1], nil
    }
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
func (lr *ListRecordsIterator) Next() (OaipmhRecord, error) {
    if (lr.pos >= len(lr.records)) {
        if (lr.resToken == "") {
            return OaipmhRecord{}, ENoMore{}
        }

        err := lr.fetchNext()
        if (err != nil) {
            return OaipmhRecord{}, err
        } else {
            return lr.Next()
        }
    } else {
        lr.pos++
        return lr.records[lr.pos - 1], nil
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
