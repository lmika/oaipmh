// oaipmh-parser.go - The OAI-PMH Parser

package main

import (
    "fmt"
    "os"
    "bytes"
    "net/url"
    "net/http"
    "time"

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
    Set             string
    Deleted         bool
}


// The result from listing a set
type ListSetResult struct {
    Spec            string
    Name            string
}


// =================================================================================
// The Oaipmh Session


type OaipmhSession struct {
    url         string
    prefix      string
    urlTraceFn  func(string)
}

// Creates a new OaipmhSession
func NewOaipmhSession(url, prefix string) *OaipmhSession {
    return &OaipmhSession{url, prefix, nil}
}

// Sets the function to call when fetching a URL
func (op *OaipmhSession) SetUrlTraceFunction(fn func(string)) {
    op.urlTraceFn = fn
}

// Gets a request from the OAI-PMH provider and returns it as a string, or an error
func (op *OaipmhSession) request(verb string, args url.Values) ([]byte, error) {
    args.Add("verb", verb)

    if (op.urlTraceFn != nil) {
        traceUrl := op.url + "?" + args.Encode()
        op.urlTraceFn(traceUrl)
    }
    
    resp, err := http.PostForm(op.url, args)
    if err != nil {
        return nil, err
    }

    respData := bytes.Buffer{}
    respData.ReadFrom(resp.Body)

    return respData.Bytes(), nil
}

// Like request but returns the result as an XML document
func (op *OaipmhSession) requestXml(verb string, args url.Values) xml.Document {
    resp, err := op.request(verb, args)
    if err != nil {
        panic(err)
    }

    doc, err := gokogiri.ParseXml(resp)
    if err != nil {
        panic(err)
    }

    return doc
}

// Runs an XPath.  Returns a slice of nodes.
func (op *OaipmhSession) runXPath(doc xml.Document, expr string) []xml.Node {
    xpathExpr := xpath.Compile(expr)
    defer xpathExpr.Free()

    xpath := xpath.NewXPath(doc.DocPtr())
    xpath.RegisterNamespace("o", "http://www.openarchives.org/OAI/2.0/")
    xpath.RegisterNamespace("gmd", "http://www.isotc211.org/2005/gmd")

    resultNodes, err := xpath.Evaluate(doc.Root().NodePtr(), xpathExpr)
    if err != nil {
        panic(err)
    }

    nodes := make([]xml.Node, len(resultNodes))
    for i, rp := range resultNodes {
        nodes[i] = xml.NewNode(rp, doc)
    }

    return nodes
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
func (op *OaipmhSession) requestXmlList(verb string, args url.Values, xpath string, firstResult int, maxResults int, callback func(node xml.Node) bool) {

    var resultCount int = 0

    for {
        // Make the request
        doc := op.requestXml(verb, args)

        // Extract the nodes
        nodes := op.runXPath(doc, xpath)

        // Run the callback for each node
        for _, n := range nodes {
            if (resultCount >= firstResult) {
                if (! callback(n)) {
                    return
                }
            }
            resultCount++
            if ((resultCount >= firstResult + maxResults) && (maxResults != -1)) {
                fmt.Fprintf(os.Stderr, "Maximum number of results encountered (%d).  Use -c to change.\n", maxResults)
                return
            }
        }

        // If there is a resumption token, use it and make the next request
        res := op.runXPath(doc, "/o:OAI-PMH//o:resumptionToken")
        if (len(res) == 1) && (op.safeNodeContents(res[0]) != "") {
            args = url.Values {
                "resumptionToken": {op.safeNodeContents(res[0])},
            }
        } else {
            return
        }
    }
}


// Returns a list of identifiers
func (op *OaipmhSession) ListIdentifiers(listArgs ListIdentifierArgs, firstResult int, maxResults int, callback func(res ListIdentifierResult) bool) {
    args := url.Values {
        "metadataPrefix":   {op.prefix},
    }
    if (listArgs.From != nil) {
        args.Add("from", listArgs.From.UTC().Format(time.RFC3339))
    }
    if (listArgs.Until != nil) {
        args.Add("until", listArgs.Until.UTC().Format(time.RFC3339))
    }

    xpath := "/o:OAI-PMH/o:ListIdentifiers/o:header"

    // Set additional arguments
    if (listArgs.Set != "") {
        args.Set("set", listArgs.Set)
    }

    op.requestXmlList("ListIdentifiers", args, xpath, firstResult, maxResults, func(node xml.Node) bool {
        id := op.safeNodeContents(op.findChild(node, "identifier"))
        dateStamp := op.safeNodeContents(op.findChild(node, "datestamp"))
        set := op.safeNodeContents(op.findChild(node, "setSpec"))
        isDeleted := (node.Attr("status") == "deleted")

        return callback(ListIdentifierResult{id, dateStamp, set, isDeleted})
    })
}

// Lists the sets provided by this provider
func (op *OaipmhSession) ListSets(firstResult int, maxResults int, callback func(ListSetResult) bool) {
    args := url.Values {}
    xpath := "/o:OAI-PMH/o:ListSets/o:set"

    op.requestXmlList("ListSets", args, xpath, firstResult, maxResults, func(node xml.Node) bool {
        spec := op.safeNodeContents(op.findChild(node, "setSpec"))
        name := op.safeNodeContents(op.findChild(node, "setName"))

        return callback(ListSetResult{spec, name})
    })
}


// Returns the record as a string
func (op *OaipmhSession) GetRecord(id string) string {
    args := url.Values{
        "metadataPrefix":   {op.prefix},
        "identifier":       {id},
    }

    doc := op.requestXml("GetRecord", args)
    return doc.String()
}

// Returns the record header as an array of string pairs
func (op *OaipmhSession) GetRecordHeader(id string) [][]string {
    args := url.Values{
        "metadataPrefix":   {op.prefix},
        "identifier":       {id},
    }

    doc := op.requestXml("GetRecord", args)

    // Parse the XML document
    headerNodes := op.runXPath(doc, "/o:OAI-PMH/o:GetRecord/o:record/o:header/*")
    res := make([][]string, len(headerNodes))

    for i, h := range(headerNodes) {
        res[i] = []string { h.Name(), h.Content() }
    }

    return res
}

// Returns the record payload as a string
func (op *OaipmhSession) GetRecordPayload(id string) string {
    args := url.Values{
        "metadataPrefix":   {op.prefix},
        "identifier":       {id},
    }

    doc := op.requestXml("GetRecord", args)

    metadataNodes := op.runXPath(doc, "/o:OAI-PMH/o:GetRecord/o:record/o:metadata/*")
    if (len(metadataNodes) == 0) {
        return ""
    } else if (len(metadataNodes) == 1) {
        return metadataNodes[0].String()
    } else {
        panic(fmt.Sprintf("Expected exactly one child of 'metadata', but found %d", len(metadataNodes)))
    }
}
