// Response structs used by the HTTP handler.
//

package oaipmh

import (
    "time"
    "encoding/xml"
)


type OaipmhResponsePayload interface{}


type OaipmhResponse struct {
    XMLName         xml.Name                `xml:"http://www.openarchives.org/OAI/2.0/ OAI-PMH"`
    Date            time.Time               `xml:"responseDate"`
    Request         OaipmhResponseRequest   `xml:"request"`

    // Possible response types.  Only ONE of these may be set at one type
    Error                   *OaipmhError
    Identify                *OaipmhIdentify
    ListMetadataFormats     *OaipmhListMetadataFormats
    ListSets                *OaipmhListSets
    ListIdentifiers         *OaipmhListIdentifiers
    ListRecords             *OaipmhListRecords
    GetRecord               *OaipmhGetRecord
}

func (res *OaipmhResponse) SetPayload(p OaipmhResponsePayload) {
    switch r := p.(type) {
        case *OaipmhError:
            res.Error = r
        case *OaipmhIdentify:
            res.Identify = r
        case *OaipmhListMetadataFormats:
            res.ListMetadataFormats = r
        case *OaipmhListSets:
            res.ListSets = r
        case *OaipmhListIdentifiers:
            res.ListIdentifiers = r
        case *OaipmhListRecords:
            res.ListRecords = r
        case *OaipmhGetRecord:
            res.GetRecord = r
        default:
            panic("Invalid oaipmh response payload")
    }
}


type OaipmhResponseRequest struct {
    Host            string                  `xml:",chardata"`
    Verb            string                  `xml:"verb,attr"`
}

// Payload for an error
type OaipmhError struct {
    XMLName         xml.Name                `xml:"error"`
    Code            string                  `xml:"code,attr"`
    Message         string                  `xml:",chardata"`
}

// Payload for returning the identity of this repository
type OaipmhIdentify struct {
    XMLName         xml.Name                `xml:"Identify"`
    RepositoryName  string                  `xml:"repositoryName"`
    BaseURL         string                  `xml:"baseUrl"`
    ProtocolVer     string                  `xml:"protocolVersion"`
    AdminEmail      string                  `xml:"adminEmail"`
    EarliestDatestamp string                `xml:"earliestDatestamp"`
    DeletedRecord   string                  `xml:"deletedRecord"`
    Granularity     string                  `xml:"granularity"`
}

// Payload for a list of formats
type OaipmhListMetadataFormats struct {
    XMLName         xml.Name                `xml:"ListMetadataFormats"`
    Formats         []Format                `xml:"metadataFormat"`
}

// Payload for listing sets
type OaipmhListSets struct {
    XMLName         xml.Name                `xml:"ListSets"`
    Sets            []OaipmhSet             `xml:"set"`
}

// Payload for listing identifiers
type OaipmhListIdentifiers struct {
    XMLName         xml.Name                `xml:"ListIdentifiers"`
    Headers         []OaipmhHeader          `xml:"header"`
    ResumptionToken string                  `xml:"resumptionToken,omitempty"`
}

// Payload for listing records
type OaipmhListRecords struct {
    XMLName         xml.Name                `xml:"ListRecords"`
    Records         []OaipmhRecord          `xml:"record"`
    ResumptionToken string                  `xml:"resumptionToken,omitempty"`
}

// Header
type OaipmhHeader struct {
    Identifier      string                  `xml:"http://www.openarchives.org/OAI/2.0/ identifier"`
    DateStamp       time.Time               `xml:"http://www.openarchives.org/OAI/2.0/ datestamp"`
    SetSpec         []string                `xml:"http://www.openarchives.org/OAI/2.0/ setSpec"`
    Status          string                  `xml:"status,attr"`
}

func RecordToOaipmhHeader(rec *Record) OaipmhHeader {
    return OaipmhHeader{
        Identifier: rec.ID,
        DateStamp: rec.Date.In(time.UTC),
        SetSpec: []string { rec.Set },
    }
}

// Record
type OaipmhRecord struct {
    XMLName         xml.Name                `xml:"record"`
    Header          OaipmhHeader            `xml:"header"`
    Content         OaipmhContent           `xml:"metadata"`
}

type OaipmhContent struct {
    Xml             string                  `xml:",innerxml"`
}

func RecordToOaipmhRecord(rec *Record) (OaipmhRecord, error) {
    content, err := rec.Content()
    if (err != nil) {
        return OaipmhRecord{}, err
    } else {
        return OaipmhRecord{
            Header: RecordToOaipmhHeader(rec),
            Content: OaipmhContent{content},
        }, err
    }
}

// GetRecord
type OaipmhGetRecord struct {
    XMLName         xml.Name                `xml:"GetRecord"`
    Record          OaipmhRecord            `xml:"record"`
}

// Single set for listing
type OaipmhSet struct {
    Spec        string                      `xml:"setSpec"`
    Name        string                      `xml:"setName"`
    Descr       OaipmhSetDescr              `xml:"setDescription"`
}

type OaipmhSetDescr struct {
    OaiDC       OaipmhOaiDC                 `xml:"http://www.openarchives.org/OAI/2.0/oai_dc/ dc"`
}

type OaipmhOaiDC struct {
    Descr       string                      `xml:"description"`
}
