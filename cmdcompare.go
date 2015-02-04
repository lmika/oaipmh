package main

import (
    "log"
    "flag"
    "os"
    "fmt"
)

// --------------------------------------------------------------------------------
// CompareWith command
//      Compares the set of metadata records in one catalog with the records of
//      another.

type CompareCommand struct {
    Ctx                 *Context
    OtherProvider       *Provider
    OtherSession        *OaipmhSession

    setName             *string
    beforeDate          *string
    afterDate           *string
    fromFile            *string
    firstResult         *int
    maxResults          *int

    compareContent      *bool

    urnsInBoth          int
    missingUrns         int
    redundentUrns       int
    urnsDiffering       int
    errors              int
}

// Startup flags
func (sc *CompareCommand) Flags(fs *flag.FlagSet) *flag.FlagSet {
    sc.setName = fs.String("s", "\x00", "Select records from this set")
    sc.beforeDate = fs.String("B", "", "Select records that were updated before date (YYYY-MM-DD)")
    sc.afterDate = fs.String("A", "", "Select records that were updated after date (YYYY-MM-DD)")
    sc.firstResult = fs.Int("f", 0, "Index of first record to retrieve")
    sc.fromFile = fs.String("F", "", "Read identifiers from a file")
    sc.maxResults = fs.Int("c", 100000, "Maximum number of records to retrieve")
    sc.compareContent = fs.Bool("C", false, "Compares the metadata content of common metadata records")

    return fs
}


// Get list identifier arguments
func (sc *CompareCommand) genListIdentifierArgsFromCommandLine() ListIdentifierArgs {
    var set string

    set = *(sc.setName)
    if set == "\x00" {
        set = sc.Ctx.Provider.Set
    }

    args := ListIdentifierArgs{
        Set: set,
        From: parseDateString(*(sc.afterDate)),
        Until: parseDateString(*(sc.beforeDate)),
    }

    return args
}

// Returns a suitable lister for the expected comparator
func (sc *CompareCommand) expectedLister() PresenceLister {
    if *(sc.fromFile) != "" {
        return func(callback func(urn string, isLive bool) bool) error {
            return LinesFromFile(*(sc.fromFile), *(sc.firstResult), *(sc.maxResults), func(urn string) bool {
                return callback(urn, true)
            })
        }
    } else {
        listArgs := sc.genListIdentifierArgsFromCommandLine()
        return func(callback func(urn string, isLive bool) bool) error {
            return sc.Ctx.Session.ListIdentifiers(listArgs, *(sc.firstResult), *(sc.maxResults), func(hr *HeaderResult) bool {
                return callback(hr.Identifier(), !hr.Deleted)
            })
        }
    }
}

// Returns a suitable lister for the comparison endpoint
func (sc *CompareCommand) comparisonLister() PresenceLister {
    // TODO: When getting it from file, simply get headers from the other session
    if *(sc.fromFile) != "" {
        Die("Support for file not done yet")
        return nil
    } else {
        listArgs := sc.genListIdentifierArgsFromCommandLine()
        return func(callback func(urn string, isLive bool) bool) error {
            return sc.OtherSession.ListIdentifiers(listArgs, *(sc.firstResult), *(sc.maxResults), func(hr *HeaderResult) bool {
                return callback(hr.Identifier(), !hr.Deleted)
            })
        }
    }
}

// Runs the presence comparator
func (sc *CompareCommand) runPresenceComparator() {
    pc := InMemoryPresenceComparator(make(map[string]byte))

    // Run the expected lister
    expectedLister := sc.expectedLister()
    expectedLister(func(urn string, isLive bool) bool {
        if (isLive) {
            pc.AddExpectedUrn(urn)
        }
        return true
    })

    // Runs the comparison lister
    comparisonLister := sc.comparisonLister()
    comparisonLister(func(urn string, isLive bool) bool {
        if (isLive) {
            pc.AddComparisonUrn(urn)
        }
        return true
    })

    // Return the report
    pc.Report(sc)
}

// Called by the presence comparison lister with URNs that are present in both providers.
func (sc *CompareCommand) UrnPresentInBothProviders(urn string) {
    sc.urnsInBoth++

    // Compare both records if in comparison mode
    if *sc.compareContent {
        thisRec, err := sc.Ctx.Session.GetRecord(urn)
        if err != nil {
            fmt.Println("E ", urn)
            sc.errors++
        }

        otherRec, err := sc.OtherSession.GetRecord(urn)
        if err != nil {
            fmt.Println("E ", urn)
            sc.errors++
        }

        if thisRec.Content.Xml != otherRec.Content.Xml {
            fmt.Println("D ", urn)
            sc.urnsDiffering++
        }
    }
}

// Called by the presence comparison lister with URNs that is in the expected provider but missing
// from the comparison provider.
func (sc *CompareCommand) MissingUrnFound(urn string) {
    fmt.Println("- ", urn)
    sc.missingUrns++
}

// Called by the presence comparison lister with URNs that are in the comparison provider but missing
// from the expected provider.
func (sc *CompareCommand) RedundentUrnFound(urn string) {
    fmt.Println("+ ", urn)
    sc.redundentUrns++
}

// Runs the comparator
func (sc *CompareCommand) Run(args []string) {
    if (len(args) != 1) {
        fmt.Fprintf(os.Stderr, "Usage: compare <provider>\n")
        os.Exit(1)
    }

    // Connect to the other OAIPMH session
    sc.OtherProvider = sc.Ctx.Config.LookupProvider(args[0])
    if (sc.OtherProvider != nil) {
        sc.OtherSession = NewOaipmhSession(sc.OtherProvider.Url, *prefix)
    } else {
        Die("Could not log into provider %s", args[0])
    }

    // Runs the presence comparator
    sc.runPresenceComparator()

    if *sc.compareContent {
        log.Printf("Comparison complete: %d OK, %d different, %d missing, %d redundent, %d errors", 
                sc.urnsInBoth, sc.urnsDiffering, sc.missingUrns, sc.redundentUrns, sc.errors)
    } else {
        log.Printf("Comparison complete: %d OK, %d missing, %d redundent", sc.urnsInBoth, sc.missingUrns, sc.redundentUrns)
    }
}


// -------------------------------------------------------------------------------------
// Maintains the comparison state

type PresenceLister     func(callback func(urn string, islive bool) bool) error

type PresenceComparator interface {

    // Adds a URN from the "expected" provider.
    AddExpectedUrn(urn string)

    // Adds a URN from the "comparison" provider.
    AddComparisonUrn(urn string)

    // Report the results
    Report(listener PresenceComparisonStateListener)
}

// Listener which is used by the comparison state to report differences.
type PresenceComparisonStateListener interface {

    // Called with URNs that are present in both providers.
    UrnPresentInBothProviders(urn string)

    // Called with URNs that is in the expected provider but missing
    // from the comparison provider.
    MissingUrnFound(urn string)

    // Called with URNs that are in the comparison provider but missing
    // from the expected provider.
    RedundentUrnFound(urn string)
}

// Bitmasks for the state of a single URN
const (
    URN_IN_EXPECTED     byte    =   0x01
    URN_IN_ACTUAL       byte    =   0x02
    URN_IN_BOTH         byte    =   0x03
)

// A presence comparator that uses an in-memory map
type InMemoryPresenceComparator map[string]byte

func (mpc InMemoryPresenceComparator) setBitForUrn(urn string, theBit byte) {
    if currVal, hasUrn := mpc[urn] ; hasUrn {
        mpc[urn] = currVal | theBit
    } else {
        mpc[urn] = theBit
    }
}

func (mpc InMemoryPresenceComparator) AddExpectedUrn(urn string) {
    mpc.setBitForUrn(urn, URN_IN_EXPECTED)
}

func (mpc InMemoryPresenceComparator) AddComparisonUrn(urn string) {
    mpc.setBitForUrn(urn, URN_IN_ACTUAL)
}

func (mpc InMemoryPresenceComparator) Report(listener PresenceComparisonStateListener) {
    for urn, bitmask := range mpc {
        switch bitmask {
            case URN_IN_BOTH:
                listener.UrnPresentInBothProviders(urn)
            case URN_IN_EXPECTED:
                listener.MissingUrnFound(urn)
            case URN_IN_ACTUAL:
                listener.RedundentUrnFound(urn)
            default:
                panic(fmt.Errorf("Invalid value of bitmask: %s, %x", urn, bitmask))
        }
    }
}
