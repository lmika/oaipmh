// A handler suitable for hosting a repository at a HTTP endpoint.
//

package oaipmh

import (
    "net/http"
    "encoding/xml"
    "time"
    "log"
    "fmt"
    "strings"
    "strconv"

    "github.com/nu7hatch/gouuid"
)

// Handler verb
type handlerVerb    func(req *http.Request) (OaipmhResponsePayload, error)

// A OAI-PMH handler.  This can be used to host a repository as a OAI-PMH provider.
//
type Handler struct {
    // The repostiory to host
    Repository      Repository

    // The supported verbs.  This simplifies the dispatching of requests.
    verbs           map[string]handlerVerb

    // The set of active resumption tokens mapped to IDs
    resumptionToks  map[string]*ResumptionToken
}

// Creates a new handler
func NewHandler(repo Repository) *Handler {
    h := &Handler{
        Repository: repo,
        verbs: make(map[string]handlerVerb),
        resumptionToks: make(map[string]*ResumptionToken),
    }

    // Set the verbs
    h.verbs["listmetadataformats"] = h.listMetadataFormats
    h.verbs["listsets"] = h.listSets
    h.verbs["identify"] = h.identify
    h.verbs["listidentifiers"] = h.listIdentifiers
    h.verbs["listrecords"] = h.listRecords

    return h
}


// Serves a HTTP response
func (h *Handler) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
    req.ParseForm()

    verb := req.Form.Get("verb")
    payload, err := h.dispatch(verb, req)
    if (err != nil) {
        log.Printf("Internal server error during dispatch: verb = %s, error = %s", verb, err.Error())
        http.Error(rw, err.Error(), http.StatusInternalServerError)
        return
    }

    fullResponse := OaipmhResponse{
        Date: time.Now().In(time.UTC),
        Request: OaipmhResponseRequest{
            Host: "http://" + req.Host + "/",
            Verb: verb,
        },
        Payload: payload,
    }

    s, err := xml.MarshalIndent(fullResponse, "  ", "    ")
    if (err != nil) {
        log.Printf("Internal server error while writing response: verb = %s, error = %s", verb, err.Error())
        http.Error(rw, err.Error(), http.StatusInternalServerError)
        return
    }

    // Serialise it as XML
    rw.Header().Set("Content-type", "application/xml")

    rw.Write([]byte(xml.Header + "\n"))
    rw.Write([]byte(s))
}

// Dispatches the request
func (h *Handler) dispatch(verb string, req *http.Request) (OaipmhResponsePayload, error) {
    log.Printf("Request: verb=%s", verb)

    verbHandler, hasVerbHandler := h.verbs[strings.ToLower(verb)]
    if (! hasVerbHandler) {
        return &OaipmhError{
            Code: "badVerb",
            Message: "Illegal OAI verb: " + verb,
        }, nil
    }

    return verbHandler(req)
}

// Identify the repository
func (h *Handler) identify(req *http.Request) (OaipmhResponsePayload, error) {
    return &OaipmhIdentify{
        RepositoryName: "oaipmh-viewer served repository",
        BaseURL: "http://" + req.Host + "/",
        ProtocolVer: "2.0",
        EarliestDatestamp: MinTime.In(time.UTC).Format(time.RFC3339),
        DeletedRecord: "transient",
        Granularity: "YYYY-MM-DDThh:mm:ssZ",
        AdminEmail: "",
    }, nil
}


// Lists the metadata formats
func (h *Handler) listMetadataFormats(req *http.Request) (OaipmhResponsePayload, error) {
    // Returns the slice of sets from the repository
    formats := h.Repository.Formats()

    return &OaipmhListMetadataFormats{
        Formats: formats,
    }, nil
}

// List the metadata sets
func (h *Handler) listSets(req *http.Request) (OaipmhResponsePayload, error) {
    sets, err := h.Repository.Sets()
    if (err != nil) {
        return nil, err
    }

    // Convert the sets into OAI-PMH sets.
    oaipmhSets := make([]OaipmhSet, len(sets))
    for i, set := range sets {
        oaipmhSets[i] = OaipmhSet{set.Spec, set.Name, OaipmhSetDescr{OaipmhOaiDC{set.Descr}}}
    }

    return &OaipmhListSets {
        Sets: oaipmhSets,
    }, nil
}

// List metadata identifiers
func (h *Handler) listIdentifiers(req *http.Request) (OaipmhResponsePayload, error) {
    cursor, err := h.getCursorForListVerb(req)
    if (err != nil) {
        return nil, err
    }

    // List the records.
    recs, _ := NextNRecords(cursor, 100)
    headers := make([]OaipmhHeader, len(recs))
    for i, rec := range recs {
        headers[i] = RecordToOaipmhHeader(rec)
    }

    resumptionToken, _ := h.storeCursorState(cursor)
    return &OaipmhListIdentifiers{
        Headers: headers,
        ResumptionToken: resumptionToken,
    }, nil
}

// List metadata records
func (h *Handler) listRecords(req *http.Request) (OaipmhResponsePayload, error) {
    cursor, err := h.getCursorForListVerb(req)
    if (err != nil) {
        return nil, err
    }

    // List the records.
    recs, _ := NextNRecords(cursor, 100)
    records := make([]OaipmhRecord, len(recs))
    for i, rec := range recs {
        records[i], err = RecordToOaipmhRecord(rec)
        if (err != nil) {
            return nil, err
        }
    }

    resumptionToken, _ := h.storeCursorState(cursor)
    return &OaipmhListRecords{
        Records: records,
        ResumptionToken: resumptionToken,
    }, nil
}

// Get a cursor for a list verb.
func (h *Handler) getCursorForListVerb(req *http.Request) (RecordCursor, error) {
    var cursor RecordCursor
    var err error

    if (req.Form.Get("resumptionToken") != "") {
        cursor = h.loadCursorState(req.Form.Get("resumptionToken"))
    } else {
        set := req.Form.Get("set")
        cursor, err = h.Repository.ListRecords(set, MinTime, time.Now())
        if (err != nil) {
            return nil, err
        }
    }

    return cursor, nil
}

// Store the cursor state and returns a resumption token if required.
func (h *Handler) storeCursorState(cursor RecordCursor) (string, bool) {
    if (cursor.HasRecord()) {
        rt := NewResumptionToken(cursor)
        h.resumptionToks[rt.ID] = rt
        return fmt.Sprintf("%s/%d", rt.ID, cursor.Pos()), true
    } else {
        return "", false
    }
}

// Load cursor state.  Returns nil if no resumption token was found.
func (h *Handler) loadCursorState(resumptionToken string) RecordCursor {
    var id string
    var pos int

    toks := strings.Split(resumptionToken, "/")
    if (len(toks) != 2) {
        return nil
    }
    id = toks[0]
    pos, _ = strconv.Atoi(toks[1])

    cursor := h.resumptionToks[id].Cursor
    if (cursor == nil) {
        return nil
    }
    defer delete(h.resumptionToks, id)

    cursor.SetPos(pos)

    return cursor
}

// ------------------------------------------------------------------------------
// Resumption token

type ResumptionToken struct {
    // The token ID
    ID          string

    // The time the token was created
    Created     time.Time

    // The cursor
    Cursor      RecordCursor
}

// Creates a new resumption token
func NewResumptionToken(cursor RecordCursor) *ResumptionToken {
    id, _ := uuid.NewV4()
    return &ResumptionToken{id.String(), time.Now(), cursor}
}
