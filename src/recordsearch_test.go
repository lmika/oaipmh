package main

import "testing"


// Test parsing of a search predicate
func TestParseXpath(t *testing.T) {
    res, err := ParseRecordMatchExpr(`xp("/a/b/c")`)
    if err != nil {
        t.Error(err)
    }

    xpStruct := res.(*XPathExistsMatchNode)

    if xpStruct.String() != "xp(/a/b/c)" {
        t.Error("Wrong expression: " + xpStruct.String())
    }
}

// Test the search predicate.
func TestXPMatch(t *testing.T) {
    res, err := ParseRecordMatchExpr(`xp("/a/b/c")`)
    if err != nil {
        t.Error(err)
    }

    rec1 := &RecordResult{}
    rec1.Content = "<gmd:a xmlns:gmd=\"something\"><val>Some value</val><b><d>Something</d><c>I've got C</c></b></gmd:a>"

    rec2 := &RecordResult{}
    rec2.Content = "<gmd:a xmlns:gmd=\"something\"><val>Some value</val><b><d>No C Here</d></b></gmd:a>"

    if v, r, _ := res.Match(rec1) ; !(r && (v=="I've got C")) {
        t.Error("rec1 must match")
    }
    if _, r, _ := res.Match(rec2) ; !(!r) {
        t.Error("rec2 must not match")
    }
}
