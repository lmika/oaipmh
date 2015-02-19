package main

import (
    "testing"
    "net/url"
)

func TestEscapeIdForFilename(t *testing.T) {
    assertEscapedIdIsEqual(t, "", EscapeIdForFilename(""))

    assertEscapedIdIsEqual(t, "abc123", ("abc123"))
    assertEscapedIdIsEqual(t, "urn:x-xmo:wis::int.wmo.wis:SSVX13", ("urn:x-xmo:wis::int.wmo.wis:SSVX13"))
    assertEscapedIdIsEqual(t, "All_valid-Characters-4_this.test", ("All_valid-Characters-4_this.test"))
    assertEscapedIdIsEqual(t, "some.file.xml", ("some.file.xml"))

    assertEscapedIdIsEqual(t, "spaces%20are%20here", ("spaces are here"))
    assertEscapedIdIsEqual(t, "abc%2F123", ("abc/123"))
    assertEscapedIdIsEqual(t, "someone%40somewhere%23here", ("someone@somewhere#here"))
    assertEscapedIdIsEqual(t, "%7E%60%21%40%23%24%25%5E%26%2A%28%29_%2B-%3D%5B%5D%7B%7D%5C%7C:%3B%22%27%3C%3E%2C.%3F%2F",
        ("~`!@#$%^&*()_+-=[]{}\\|:;\"'<>,.?/"))
}


func assertEscapedIdIsEqual(t *testing.T, exp, actual string) {
    // Test escaping
    if exp != EscapeIdForFilename(actual) {
        t.Errorf("assertEscapedIdIsEqual: escaping - expected '%s' but received '%s'", exp, actual)
    }

    // Test unescaping
    unescape, err := url.QueryUnescape(exp)
    if err != nil {
        t.Error(err)
    } else if unescape != actual {
        t.Errorf("assertEscapedIdIsEqual: unescaping - expected '%s' but received '%s'", actual, unescape)
    }
}