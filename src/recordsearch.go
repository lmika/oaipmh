// Record search.  Parses a search expression and runs it over a RecordResult returning true if the
// record matches the expression, or false otherwise.
//
//  Expressions are of the form:
//
//      xp( <string> )          -- Run the XPath over the contents of the record, returning the
//                                 boolean value of the xpath.
//

package main

import (
    "text/scanner"
    "strconv"
    "strings"
    "fmt"

    "launchpad.net/xmlpath"
)

// Node for a record match.
type RecordMatchNode interface {
    // Returns true if the particular matches the record node.
    Match(rr *RecordResult) (string, bool, error)

    // Returns a string implementation of the record match.
    String() string
}


// ------------------------------------------------------------------------------
// xp(<xpath>)
//      Returns true if the XPath exists.

type XPathExistsMatchNode struct {
    path        *xmlpath.Path
    origExpr    string
}

func (xe *XPathExistsMatchNode) Match(rr *RecordResult) (string, bool, error) {
    n, err := xmlpath.Parse(strings.NewReader(rr.Content))
    if (err != nil) {
        return "", false, err
    }

    val, hasVal := xe.path.String(n)

    return val, hasVal, nil
}

func (xe *XPathExistsMatchNode) String() string {
    return "xp(" + xe.origExpr + ")"
}


// ------------------------------------------------------------------------------
//

// Record search parser
type recordSearchParser struct {
    scan        *scanner.Scanner
    tok         rune
    tokText     string
}

// Gets the next token
func (rsp *recordSearchParser) nextToken() {
    if (rsp.tok != scanner.EOF) {
        rsp.tok = rsp.scan.Scan()
        rsp.tokText = rsp.scan.TokenText()
    }
}

// Consumes a token.  Returns the token value
func (rsp *recordSearchParser) consume(tok rune) (txt string, err error) {
    if (rsp.tok == tok) {
        txt = rsp.tokText
        rsp.nextToken()
    } else {
        err = fmt.Errorf("Expected %s but got %s\n", scanner.TokenString(tok), scanner.TokenString(rsp.tok))
    }
    return
}

// Parses a function call
func (rsp *recordSearchParser) parseFn() (RecordMatchNode, error) {
    fnName, err := rsp.consume(scanner.Ident)
    if (err != nil) {
        return nil, err
    }

    if _, err = rsp.consume('(') ; err != nil {
        return nil, err
    }

    fnArg, err := rsp.readString()
    if (err != nil) {
        return nil, err
    }
    rsp.consume(rsp.tok)

    if _, err = rsp.consume(')') ; err != nil {
        return nil, err
    }

    return rsp.buildFnNode(fnName, fnArg)
}

// Construct the function call
func (rsp *recordSearchParser) buildFnNode(name string, arg string) (RecordMatchNode, error) {
    if (name == "xp") {
        path, err := xmlpath.Compile(arg)
        if (err != nil) {
            return nil, err
        }
        return &XPathExistsMatchNode{path, arg}, nil
    } else {
        return nil, fmt.Errorf("Unknown search function: %s\n", name)
    }
}

// Reads a string value
func (rsp *recordSearchParser) readString() (string, error) {
    if (rsp.tok == scanner.String) || (rsp.tok == scanner.RawString) {
        return strconv.Unquote(rsp.tokText)
    } else {
        rsp.consume(rsp.tok)
        return rsp.tokText, nil
    }
}

// Parses a record match expression
func ParseRecordMatchExpr(expr string) (RecordMatchNode, error) {
    rsp := &recordSearchParser{}
    rsp.scan = new(scanner.Scanner)
    rsp.scan.Init(strings.NewReader(expr))
    rsp.scan.Mode = scanner.ScanIdents | scanner.ScanStrings | scanner.ScanRawStrings | scanner.SkipComments
    rsp.nextToken()

    return rsp.parseFn()
}
