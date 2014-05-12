// oaipmh-parser.go - The OAI-PMH Parser

package main

import (
    "fmt"
    "os"
    "bytes"
    "net/url"
    "net/http"
    "time"

    "./oaipmh"

    "github.com/moovweb/gokogiri"
    "github.com/moovweb/gokogiri/xml"
    "github.com/moovweb/gokogiri/xpath"
)


// The arguments to the list identifier string
type ListIdentifierArgs struct {
    Set             string          // The set to query
    From            *time.Time      // The from time (nil == no check)
    Until           *time.Time      // The until time (nil == no check)
}


// The result from listing the identifiers
type ListIdentifierResult struct {
    Identifier      string
    Datestamp       string
    Sets            []string
    Deleted         bool
}


// The result from listing a set
type ListSetResult struct {
    Spec            string
    Name            string
    Description     string
}

// Results from reading a record
type RecordResult struct {
    Header          oaipmh.OaipmhHeader
    Content         string
    Deleted         bool
}

// A record error (which can contain another error)
type RecordError struct {
    Id              string
    Err             string
}

func (r *RecordError) Error() string {
    return "Record '" + r.Id + "': " + r.Err
}


// Returns the identifier of the record.  This uses the header "identifier" field.
func (rr *RecordResult) Identifier() string {
    return rr.Header.Identifier
}

// Runs an XPath expression over the content.  This expression expects a boolean result.
func (rr *RecordResult) RunXPath(expr string) (bool, error) {
    doc, err := gokogiri.ParseXml([]byte(rr.Content))
    if err != nil {
        return false, err
    }

    xpe := xpath.Compile(expr)
    defer xpe.Free()

    xp := xpath.NewXPath(doc.DocPtr())
    xp.RegisterNamespace("o", "http://www.openarchives.org/OAI/2.0/")
    xp.RegisterNamespace("gmd", "http://www.isotc211.org/2005/gmd")
    xp.RegisterNamespace("gco", "http://www.isotc211.org/2005/gco")
    xp.RegisterNamespace("gml", "http://www.opengis.net/gml/3.2")
    xp.RegisterNamespace("gts", "http://www.isotc211.org/2005/gts")
    xp.RegisterNamespace("dc", "http://www.openarchives.org/OAI/2.0/oai_dc/")

    err = xp.Evaluate(doc.Root().NodePtr(), xpe)
    if err != nil {
        panic(err)
    }

    b, err := xp.ResultAsBoolean()
    return b, err
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

// Sets the function to call when fetching a URL
func (op *OaipmhSession) SetUrlTraceFunction(fn func(string)) {
    op.traceFn = fn
}

// Gets a request from the OAI-PMH provider and returns it as a string, or an error
func (op *OaipmhSession) request(verb string, args url.Values) ([]byte, error) {
    args.Add("verb", verb)

    traceUrl := op.url + "?" + args.Encode()
    op.traceFn("POST " + traceUrl)

    resp, err := http.PostForm(op.url, args)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    // If the response is not 200, return an error
    if (resp.StatusCode != 200) {
        return nil, fmt.Errorf("HTTP error: %s", resp.Status)
    }

    respData := bytes.Buffer{}
    respData.ReadFrom(resp.Body)

    return respData.Bytes(), nil
}

// Like request but returns the result as an XML document
func (op *OaipmhSession) requestXml(verb string, args url.Values) (xml.Document, error) {
    resp, err := op.request(verb, args)
    if err != nil {
        return nil, err
    }

    doc, err := gokogiri.ParseXml(resp)
    if err != nil {
        return nil, err
    }

    return doc, nil
}

// Like requestXml but recognises OAI-PMH errors.  If an error is encountered, returns an error
func (op *OaipmhSession) requestOaipmhXml(verb string, args url.Values) (xml.Document, error) {
    doc, err := op.requestXml(verb, args)
    if err != nil {
        return nil, err
    }

    // Search for an error node
    errorNode := op.runXPathSingle(doc.Root(), "/o:OAI-PMH/o:error")
    if (errorNode != nil) {
        return nil, fmt.Errorf("OAI-PMH error: %s", op.safeNodeContents(errorNode))
    }

    return doc, nil
}

// Runs an XPath.  Returns a slice of nodes.
func (op *OaipmhSession) runXPath(doc xml.Document, expr string) []xml.Node {
    xpathExpr := xpath.Compile(expr)
    defer xpathExpr.Free()

    xpath := xpath.NewXPath(doc.DocPtr())
    xpath.RegisterNamespace("o", "http://www.openarchives.org/OAI/2.0/")
    xpath.RegisterNamespace("gmd", "http://www.isotc211.org/2005/gmd")
    xpath.RegisterNamespace("dc", "http://www.openarchives.org/OAI/2.0/oai_dc/")

    resultNodes, err := xpath.EvaluateAsNodeset(doc.Root().NodePtr(), xpathExpr)
    if err != nil {
        panic(err)
    }

    nodes := make([]xml.Node, len(resultNodes))
    for i, rp := range resultNodes {
        nodes[i] = xml.NewNode(rp, doc)
    }

    return nodes
}

// Run an XPath returning a single node.
func (op *OaipmhSession) runXPathSingle(node xml.Node, expr string) xml.Node {
    xpathExpr := xpath.Compile(expr)
    defer xpathExpr.Free()

    xpath := xpath.NewXPath(node.MyDocument().DocPtr())
    xpath.RegisterNamespace("o", "http://www.openarchives.org/OAI/2.0/")
    xpath.RegisterNamespace("gmd", "http://www.isotc211.org/2005/gmd")
    xpath.RegisterNamespace("dc", "http://www.openarchives.org/OAI/2.0/oai_dc/")

    resultNodes, err := xpath.EvaluateAsNodeset(node.NodePtr(), xpathExpr)
    if err != nil {
        panic(err)
    }

    if (len(resultNodes) == 1) {
        return xml.NewNode(resultNodes[0], node.MyDocument())
    } else if (len(resultNodes) == 0) {
        return nil
    } else {
        panic("Got more than one node from XPath '" + expr + "'")
    }
}

// Searches for a child node based on the node name.
func (op *OaipmhSession) findChild(node xml.Node, name string) xml.Node {
    for n := node.FirstChild(); n != nil; n = n.NextSibling() {
        if (n.Name() == name) {
            return n
        }
    }
    return nil
}
// Searches for the first child element
func (op *OaipmhSession) findFirstElement(node xml.Node) xml.Node {
    for n := node.FirstChild(); n != nil; n = n.NextSibling() {
        if (n.NodeType() == xml.XML_ELEMENT_NODE) {
            return n
        }
    }
    return nil
}

// Runs a function over each children with a specific name
func (op *OaipmhSession) eachChildOfName(node xml.Node, name string, fn func(child xml.Node)) {
    for n := node.FirstChild(); n != nil; n = n.NextSibling() {
        if (n.Name() == name) {
            fn(n)
        }
    }
}

// Gets the contents of a node safely.
func (op *OaipmhSession) safeNodeContents(node xml.Node) string {
    if (node != nil) {
        return node.Content()
    } else {
        return ""
    }
}

// Performs a list request.  This retrieves a list of items and returns each item matching the XPath expression.
// If a resumption token is present, and the callback continues to return true, the next set of items is retrieved.
func (op *OaipmhSession) requestXmlList(verb string, args url.Values, xpath string, firstResult int, maxResults int, callback func(node xml.Node) bool) error {

    var resultCount int = 0

    for {
        // Make the request
        doc, err := op.requestOaipmhXml(verb, args)
        if err != nil {
            return err
        }

        // Extract the nodes
        nodes := op.runXPath(doc, xpath)

        // Run the callback for each node
        for _, n := range nodes {
            if (resultCount >= firstResult) {
                if (! callback(n)) {
                    return nil
                }
            }
            resultCount++
            if ((resultCount >= firstResult + maxResults) && (maxResults != -1)) {
                fmt.Fprintf(os.Stderr, "Maximum number of results encountered (%d).  Use -c to change.\n", maxResults)
                return nil
            }
        }

        // If there is a resumption token, use it and make the next request
        res := op.runXPath(doc, "/o:OAI-PMH//o:resumptionToken")
        if (len(res) == 1) && (op.safeNodeContents(res[0]) != "") {
            args = url.Values {
                "resumptionToken": {op.safeNodeContents(res[0])},
            }
        } else {
            return nil
        }
    }
}

// Runs an iterator function only calling the stepFn between firstResult and maxResults.  If the stepFn returns false,
// stops the iterator early.
func (op *OaipmhSession) iteratorSubset(firstResult int, maxResults int, nextFn func() (interface{}, error), stepFn func(interface{}) bool) error {
    var resultCount int = 0
    var err error
    for v, err := nextFn() ; err == nil ; v, err = nextFn() {
        if (resultCount >= firstResult) {
            if (! stepFn(v)) {
                return nil
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
func (op *OaipmhSession) ListIdentifiers(listArgs ListIdentifierArgs, firstResult int, maxResults int, callback func(res ListIdentifierResult) bool) error {
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

    err = op.iteratorSubset(firstResult, maxResults, func() (interface{}, error) {
        return ri.Next()
    }, func(v interface{}) bool {
        h := v.(oaipmh.OaipmhHeader)
        res := ListIdentifierResult{ h.Identifier, h.DateStamp.String(), h.SetSpec, (h.Status == "deleted") }
        return callback(res)
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

    err = op.iteratorSubset(firstResult, maxResults, func() (interface{}, error) {
        return ri.Next()
    }, func(v interface{}) bool {
        h := v.(oaipmh.OaipmhRecord)
        res := RecordResult{ h.Header, h.Content.Xml, h.Header.Status == "deleted" }
        return callback(&res)
    })

    if _, isNoRes := err.(oaipmh.ENoMore) ; isNoRes {
        return nil
    } else {
        return err
    }
/*
    args := url.Values {
        "metadataPrefix":   {op.prefix},
    }
    if (listArgs.From != nil) {
        args.Add("from", listArgs.From.UTC().Format(time.RFC3339))
    }
    if (listArgs.Until != nil) {
        args.Add("until", listArgs.Until.UTC().Format(time.RFC3339))
    }

    xpath := "/o:OAI-PMH/o:ListRecords/o:record"

    // Set additional arguments
    if (listArgs.Set != "") {
        args.Set("set", listArgs.Set)
    }

    return op.requestXmlList("ListRecords", args, xpath, firstResult, maxResults, func(node xml.Node) bool {
        recordResult := op.getHeaderAndMetadata(node)
        return callback(recordResult)
    })
*/
}

// Lists the sets provided by this provider
func (op *OaipmhSession) ListSets(firstResult int, maxResults int, callback func(ListSetResult) bool) error {
    args := url.Values {}
    xpath := "/o:OAI-PMH/o:ListSets/o:set"

    return op.requestXmlList("ListSets", args, xpath, firstResult, maxResults, func(node xml.Node) bool {
        spec := op.safeNodeContents(op.findChild(node, "setSpec"))
        name := op.safeNodeContents(op.findChild(node, "setName"))
        descr := op.safeNodeContents(op.findChild(node, "setDescription"))

        return callback(ListSetResult{spec, name, descr})
    })
}

// Returns the header and metadata from a record node
func (op *OaipmhSession) getHeaderAndMetadata(recordNode xml.Node) *RecordResult {
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
    panic("No longer supported")
    //return &RecordResult{headers, metadataContent, deleted}
}

// Returns the record header as an array of string pairs
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
