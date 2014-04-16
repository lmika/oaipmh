// A handler suitable for hosting a repository at a HTTP endpoint.
//

package oaipmh

import (
    "net/http"
    "encoding/xml"
    "time"
    "log"
    "strings"
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
}

// Creates a new handler
func NewHandler(repo Repository) *Handler {
    h := &Handler{
        Repository: repo,
        verbs: make(map[string]handlerVerb),
    }

    // Set the verbs
    h.verbs["listmetadataformats"] = h.listMetadataFormats
    h.verbs["listsets"] = h.listSets
    h.verbs["identify"] = h.identify

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
