package main

import "testing"


// Test parsing of a search predicate
func TestParseXpath(t *testing.T) {
    _, err := ParseRecordMatchExpr(`xp("/a/b/c")`)
    if err != nil {
        t.Error(err)
    }
}

// Test the search predicate.
func TestXPMatch(t *testing.T) {
    expr := `xp("/a/b/c")`
    rs, err := ParseRecordMatchExpr(expr)
    if err != nil {
        t.Error(err)
    }

    rec1 := &RecordResult{}
    rec1.Content = "<gmd:a xmlns:gmd=\"something\"><val>Some value</val><b><d>Something</d><c>I've got C</c></b></gmd:a>"

    rec2 := &RecordResult{}
    rec2.Content = "<gmd:a xmlns:gmd=\"something\"><val>Some value</val><b><d>No C Here</d></b></gmd:a>"

    if r, v, _ := rs.SearchRecord(rec1) ; !(r && (v=="I've got C")) {
        t.Error("rec1 must match")
    }
    if r, _, _ := rs.SearchRecord(rec2) ; !(!r) {
        t.Error("rec2 must not match")
    }
}

// Test the start with predicate
func TestStartWith(t *testing.T) {
    expr := `startsWith(xp("/a/val"), "Some")`
    rs, err := ParseRecordMatchExpr(expr)
    if err != nil {
        t.Error(err)
    }

    rec1 := &RecordResult{}
    rec1.Content = "<gmd:a xmlns:gmd=\"something\"><val>Some value</val><b><d>Something</d><c>I've got C</c></b></gmd:a>"

    rec2 := &RecordResult{}
    rec2.Content = "<gmd:a xmlns:gmd=\"something\"><val>Another value</val><b><d>No C Here</d></b></gmd:a>"

    if r, v, _ := rs.SearchRecord(rec1) ; !(r && (v=="Some value")) {
        t.Error("rec1 must match")
    }
    if r, v, _ := rs.SearchRecord(rec2) ; !(!r && (v=="")) {
        t.Error("rec2 must not match")
    }
}
