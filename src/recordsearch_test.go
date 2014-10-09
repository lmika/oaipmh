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

// Test the nonEmpty predicate
func TestNonEmptyValues(t *testing.T) {
    assertSearchExpr(t, `"This is not empty"`, "<xml></xml>", true, "This is not empty")
    assertSearchExpr(t, `"")`, "<xml></xml>", false, "")
    assertSearchExpr(t, `xp("/xml/value")`, "<xml>This has<value>a value</value> set</xml>", true, "a value")
    assertSearchExpr(t, `xp("/xml/missing")`, "<xml>This has<value>a value</value> set</xml>", false, "")
    assertSearchExpr(t, `xp("/xml/value")`, "<xml>This has<value attr=\"some attribute\" /> set</xml>", false, "")
    assertSearchExpr(t, `xp("/xml/value/@attr")`, "<xml>This has<value attr=\"some attribute\" /> set</xml>", true, "some attribute")
}


func assertSearchExpr(t *testing.T, expr string, xml string, expectedVal bool, expectedValue string) {
    rs, err := ParseRecordMatchExpr(expr)
    if err != nil {
        t.Error(err)
        return
    }

    rec1 := &RecordResult{}
    rec1.Content = xml

    r, v, _ := rs.SearchRecord(rec1)

    if r != expectedVal {
        t.Error("Result of expression must be %v but was %v", expectedVal, r)
        return
    }
    if v != expectedValue {
        t.Error("Value of expression must be %v but was %v", expectedValue, v)
        return
    }
}
