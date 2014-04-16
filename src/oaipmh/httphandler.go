// A handler suitable for hosting a repository at a HTTP endpoint.
//

package oaipmh

import (
    "net/url"
    "net/http"
    "encoding/xml"
    "time"
    "log"
    "strings"
)

// Handler verb
type handlerVerb    func(args url.Values) OaipmhResponsePayload

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

    return h
}


// Serves a HTTP response
func (h *Handler) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
    req.ParseForm()

    verb := req.Form.Get("verb")
    payload := h.dispatch(verb, req.Form)

    fullResponse := OaipmhResponse{
        Date: time.Now().In(time.UTC),
        Request: OaipmhResponseRequest{
            Host: req.URL.Host,
            Verb: verb,
        },
        Payload: payload,
    }

    s, e := xml.MarshalIndent(fullResponse, "  ", "    ")
    if (e != nil) {
        log.Printf("Internal server error: verb = %s, error = %s", verb, e.Error())
        rw.WriteHeader(http.StatusInternalServerError)
        rw.Write([]byte("Internal server error"))
        return
    }

    // Serialise it as XML
    rw.Header().Set("Content-type", "application/xml")

    rw.Write([]byte(xml.Header + "\n"))
    rw.Write([]byte(s))
}

// Dispatches the request
func (h *Handler) dispatch(verb string, args url.Values) OaipmhResponsePayload {
    log.Printf("Request: verb=%s", verb)

    verbHandler, hasVerbHandler := h.verbs[strings.ToLower(verb)]
    if (! hasVerbHandler) {
        return &OaipmhError{
            Code: "badVerb",
            Message: "Illegal OAI verb: " + verb,
        }
    }

    return verbHandler(args)
}


// Lists the metadata formats
func (h *Handler) listMetadataFormats(args url.Values) OaipmhResponsePayload {
    // Returns the slice of sets from the repository
    formats := h.Repository.Formats()

    return &OaipmhListMetadataFormats{
        Formats: formats,
    }
}
