// oaipmh-parser.go - The OAI-PMH Parser

package main

import (
    "fmt"
    "os"
    "time"

    "./oaipmh"
)


// The arguments to the list identifier string
type ListIdentifierArgs struct {
    Set             string          // The set to query
    From            *time.Time      // The from time (nil == no check)
    Until           *time.Time      // The until time (nil == no check)
}


// The result from listing the identifiers
type HeaderResult struct {
    Header          *oaipmh.OaipmhHeader
    Deleted         bool
}

func (r *HeaderResult) Identifier() string {
    return r.Header.Identifier
}


// Record result
type RecordResult struct {
    Header          oaipmh.OaipmhHeader
    Content         string
    Deleted         bool
}

func (r *RecordResult) Identifier() string {
    return r.Header.Identifier
}


// =================================================================================
// The Oaipmh Session


type OaipmhSession struct {
    client      *oaipmh.Client
    url         string
    prefix      string
    traceFn     func(string)
}

// Creates a new OaipmhSession
func NewOaipmhSession(url, prefix string) *OaipmhSession {
    c, err := oaipmh.NewClient(url)
    if (err != nil) {
        panic(err)
    }
    return &OaipmhSession{c, url, prefix, func(string) { }}
}

// Sets the debugging state
func (op *OaipmhSession) SetDebug(debug bool) {
    op.client.Debug = debug
}


// Runs an iterator function only calling the stepFn between firstResult and maxResults.  If the stepFn returns false,
// stops the iterator early.
func (op *OaipmhSession) iteratorSubset(iterator oaipmh.RecordIterator, firstResult int, maxResults int, withIter func(i oaipmh.RecordIterator) error) error {
    var resultCount int = 0
    var err error
    for err = iterator.Next() ; err == nil ; err = iterator.Next() {
        if (resultCount >= firstResult) {
            err2 := withIter(iterator)
            if (err2 != nil) {
                return err2
            }
        }

        resultCount++
        if ((resultCount >= firstResult + maxResults) && (maxResults != -1)) {
            fmt.Fprintf(os.Stderr, "Maximum number of results encountered (%d).  Use -c to change.\n", maxResults)
            return nil
        }
    }
    return err
}


// Returns a list of identifiers
func (op *OaipmhSession) ListIdentifiers(listArgs ListIdentifierArgs, firstResult int, maxResults int, callback func(res *HeaderResult) bool) error {
    var err error

    ri, err := op.client.ListIdentifiers(oaipmh.ListArgs{
        Prefix: op.prefix,
        From: listArgs.From,
        Until: listArgs.Until,
        Set: listArgs.Set,
    })
    if (err != nil) {
        return err
    }

    err = op.iteratorSubset(ri, firstResult, maxResults, func(i oaipmh.RecordIterator) error {
        h, err2 := i.Header()
        if (err2 != nil) {
            return err2
        }
        res := &HeaderResult{ h, (h.Status == "deleted") }
        if !callback(res) {
            return oaipmh.ENoMore{}
        } else {
            return nil
        }
    })

    if _, isNoRes := err.(oaipmh.ENoMore) ; isNoRes {
        return nil
    } else {
        return err
    }
}

// Returns a list of records
func (op *OaipmhSession) ListRecords(listArgs ListIdentifierArgs, firstResult int, maxResults int, callback func(recordResult *RecordResult) bool) error {
    var err error

    ri, err := op.client.ListRecords(oaipmh.ListArgs{
        Prefix: op.prefix,
        From: listArgs.From,
        Until: listArgs.Until,
        Set: listArgs.Set,
    })
    if (err != nil) {
        return err
    }

    err = op.iteratorSubset(ri, firstResult, maxResults, func(i oaipmh.RecordIterator) error {
        h, err2 := i.Record()
        if (err2 != nil) {
            return err2
        }
        res := &RecordResult{ h.Header, h.Content.Xml, h.Header.Status == "deleted" }
        if !callback(res) {
            return oaipmh.ENoMore{}
        } else {
            return nil
        }
    })

    if _, isNoRes := err.(oaipmh.ENoMore) ; isNoRes {
        return nil
    } else {
        return err
    }
}

// Lists the sets provided by this provider
func (op *OaipmhSession) ListSets(firstResult int, maxResults int, callback func(oaipmh.OaipmhSet) bool) error {
    sets, err := op.client.ListSets()
    if (err != nil) {
        return err
    }

    for _, set := range sets {
        callback(set)
    }

    return nil
}

// Returns the header and metadata from a record node
//func (op *OaipmhSession) getHeaderAndMetadata(recordNode xml.Node) *RecordResult {
    // Get the header
/*
    headerNode := op.findChild(recordNode, "header")
    headers := make([][]string, 0, headerNode.CountChildren())
    deleted := headerNode.Attr("status") == "deleted"

    for childNode := headerNode.FirstChild(); childNode != nil; childNode = childNode.NextSibling() {
        if (childNode.NodeType() == xml.XML_ELEMENT_NODE) {
            headers = append(headers , []string { childNode.Name(), childNode.Content() })
        }
    }


    // Get the metadata
    var metadataContent string
    metadataNode := op.findChild(recordNode, "metadata")
    if (metadataNode != nil) {
        // metadataContent = metadataNode.FirstChild().String()
        bufr := new(bytes.Buffer)
        bufr.WriteString("<?xml version=\"1.0\"?>\n")
        bufr.WriteString(op.findFirstElement(metadataNode).String())
        metadataContent = bufr.String()
    } else {
        metadataContent = ""
    }
*/
//    panic("No longer supported")
    //return &RecordResult{headers, metadataContent, deleted}
//}

// Returns the record header as an array of string pairs
/*
func (op *OaipmhSession) GetRecord(id string) (*RecordResult, error) {
    args := url.Values{
        "metadataPrefix":   {op.prefix},
        "identifier":       {id},
    }

    doc, err := op.requestOaipmhXml("GetRecord", args)
    if (err != nil) {
        return nil, &RecordError{id, err.Error()}
    }

    // Parse the XML document
    recordNode := op.runXPathSingle(doc.Root(), "/o:OAI-PMH/o:GetRecord/o:record")
    if (recordNode != nil) {
        return op.getHeaderAndMetadata(recordNode), nil
    } else {
        return nil, &RecordError{id, "Could not find 'record' node in entry"}
    }
}

// Returns the record payload as a string
func (op *OaipmhSession) GetRecordPayload(id string) (string, error) {
    args := url.Values{
        "metadataPrefix":   {op.prefix},
        "identifier":       {id},
    }

    doc, err := op.requestOaipmhXml("GetRecord", args)
    if (err != nil) {
        return "", &RecordError{id, err.Error()}
    }

    // Parse the XML document
    recordNode := op.runXPathSingle(doc.Root(), "/o:OAI-PMH/o:GetRecord/o:record")
    if (recordNode != nil) {
        return recordNode.String(), nil
    } else {
        return "", &RecordError{id, "Could not find 'record' node in entry"}
    }
}
*/

func (op *OaipmhSession) GetRecord(id string) (*oaipmh.OaipmhRecord, error) {
    rec, err := op.client.GetRecord(op.prefix, id)
    if (err != nil) {
        return nil, err
    } else {
        return rec, nil
    }
}
