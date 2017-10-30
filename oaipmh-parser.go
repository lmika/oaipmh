// oaipmh-parser.go - The OAI-PMH Parser

package main

import (
	"fmt"
	"os"
	"time"

	"github.com/lmika/oaipmh/client"
)

// The arguments to the list identifier string
type ListIdentifierArgs struct {
	Set   string     // The set to query
	From  *time.Time // The from time (nil == no check)
	Until *time.Time // The until time (nil == no check)
}

// The result from listing the identifiers
type HeaderResult struct {
	Header  *oaipmh.OaipmhHeader
	Deleted bool
}

func (r *HeaderResult) Identifier() string {
	return r.Header.Identifier
}

// Record result
type RecordResult struct {
	Header  oaipmh.OaipmhHeader
	Content string
	Deleted bool
}

func (r *RecordResult) AsHeaderResult() *HeaderResult {
	return &HeaderResult{&(r.Header), r.Deleted}
}

func (r *RecordResult) Identifier() string {
	return r.Header.Identifier
}

// =================================================================================
// The Oaipmh Session

type OaipmhSession struct {
	client  *oaipmh.Client
	url     string
	prefix  string
	traceFn func(string)
}

// Creates a new OaipmhSession
func NewOaipmhSession(url, prefix string) *OaipmhSession {
	c, err := oaipmh.NewClient(url)
	if err != nil {
		panic(err)
	}
	return &OaipmhSession{c, url, prefix, func(string) {}}
}

// Sets the debugging level (0 = none, 1 = request, 2 = request/response)
func (op *OaipmhSession) SetDebug(debug int) {
	if debug <= 0 {
		op.client.Debug = oaipmh.NoDebug
	} else if debug == 1 {
		op.client.Debug = oaipmh.ReqDebug
	} else if debug >= 2 {
		op.client.Debug = oaipmh.ReqRespBodyDebug
	}
}

// Set whether or not to use HTTP GET
func (op *OaipmhSession) SetUseGet(useGet bool) {
	op.client.UseGet = useGet
}

// Stifle error messages which indicate no more results.  This is so that the user doesn't see them.
func (op *OaipmhSession) stifleNoResultErrors(e error) error {
	switch err := e.(type) {
	case oaipmh.ENoMore:
		return nil
	case oaipmh.EOaipmhError:
		if err.Code == "noRecordsMatch" {
			return nil
		} else {
			return err
		}
	default:
		return err
	}
}

// Runs an iterator function only calling the stepFn between firstResult and maxResults.  If the stepFn returns false,
// stops the iterator early.
func (op *OaipmhSession) iteratorSubset(iterator oaipmh.RecordIterator, firstResult int, maxResults int, withIter func(i oaipmh.RecordIterator) error) error {
	var resultCount int = 0
	var err error
	for err = iterator.Next(); err == nil; err = iterator.Next() {
		if resultCount >= firstResult {
			err2 := withIter(iterator)
			if err2 != nil {
				return err2
			}
		}

		resultCount++
		if (resultCount >= firstResult+maxResults) && (maxResults != -1) {
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
		From:   listArgs.From,
		Until:  listArgs.Until,
		Set:    listArgs.Set,
	})
	if err != nil {
		return op.stifleNoResultErrors(err)
	}

	err = op.iteratorSubset(ri, firstResult, maxResults, func(i oaipmh.RecordIterator) error {
		h, err2 := i.Header()
		if err2 != nil {
			return err2
		}
		res := &HeaderResult{h, (h.Status == "deleted")}
		if !callback(res) {
			return oaipmh.ENoMore{}
		} else {
			return nil
		}
	})

	return op.stifleNoResultErrors(err)
}

// Returns a list of identifiers using the ListRecord verb
func (op *OaipmhSession) ListIdentifiersUsingListRecords(listArgs ListIdentifierArgs, firstResult int, maxResults int, callback func(res *HeaderResult) bool) error {
	var err error

	ri, err := op.client.ListRecords(oaipmh.ListArgs{
		Prefix: op.prefix,
		From:   listArgs.From,
		Until:  listArgs.Until,
		Set:    listArgs.Set,
	})
	if err != nil {
		return op.stifleNoResultErrors(err)
	}

	err = op.iteratorSubset(ri, firstResult, maxResults, func(i oaipmh.RecordIterator) error {
		h, err2 := i.Record()
		if err2 != nil {
			return err2
		}
		hc := new(oaipmh.OaipmhHeader)
		*hc = h.Header
		res := &HeaderResult{hc, h.Header.Status == "deleted"}
		if !callback(res) {
			return oaipmh.ENoMore{}
		} else {
			return nil
		}
	})

	return op.stifleNoResultErrors(err)
}

// Returns a list of records
func (op *OaipmhSession) ListRecords(listArgs ListIdentifierArgs, firstResult int, maxResults int, callback func(recordResult *RecordResult) bool) error {
	var err error

	ri, err := op.client.ListRecords(oaipmh.ListArgs{
		Prefix: op.prefix,
		From:   listArgs.From,
		Until:  listArgs.Until,
		Set:    listArgs.Set,
	})
	if err != nil {
		return op.stifleNoResultErrors(err)
	}

	err = op.iteratorSubset(ri, firstResult, maxResults, func(i oaipmh.RecordIterator) error {
		h, err2 := i.Record()
		if err2 != nil {
			return err2
		}
		res := RecordToRecordResult(h)
		if !callback(res) {
			return oaipmh.ENoMore{}
		} else {
			return nil
		}
	})

	return op.stifleNoResultErrors(err)
}

// Lists the sets provided by this provider
func (op *OaipmhSession) ListSets(firstResult int, maxResults int, callback func(oaipmh.OaipmhSet) bool) error {
	sets, err := op.client.ListSets()
	if err != nil {
		return err
	}

	for _, set := range sets {
		callback(set)
	}

	return nil
}

// Returns a record by ID
func (op *OaipmhSession) GetRecord(id string) (*oaipmh.OaipmhRecord, error) {
	rec, err := op.client.GetRecord(op.prefix, id)
	if err != nil {
		return nil, err
	} else {
		return rec, nil
	}
}

// Converts an OaipmhRecord into a RecordResult
func RecordToRecordResult(r *oaipmh.OaipmhRecord) *RecordResult {
	return &RecordResult{r.Header, r.Content.Xml, r.Header.Status == "deleted"}
}
