// Response structs used by the HTTP handler.
//

package oaipmh

import (
    "time"
)


type OaipmhResponse struct {
    XMLName         string                  `xml:"http://www.openarchives.org/OAI/2.0/ OAI-PMH"`
    Date            time.Time               `xml:"responseDate"`
    Request         OaipmhResponseRequest   `xml:"request"`
    Payload         OaipmhResponsePayload
}

type OaipmhResponseRequest struct {
    Host            string                  `xml:",chardata"`
    Verb            string                  `xml:"verb,attr"`
}

// Response payload
type OaipmhResponsePayload interface{}

// Payload for an error
type OaipmhError struct {
    XMLName         string                  `xml:"error"`
    Code            string                  `xml:"code,attr"`
    Message         string                  `xml:",chardata"`
}

// Payload for a list of formats
type OaipmhListMetadataFormats struct {
    XMLName         string                  `xml:"ListMetadataFormats"`
    Formats         []Format                `xml:"metadataFormat"`
}
