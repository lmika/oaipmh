/**
 * The OAIPMH Viewer
 * 
 * Tool for listing and viewing metadata records from OAIPMH providers.
 */

package main

import (
    "flag"
    "fmt"
    "os"

    "github.com/lmika/command"
)

// Flags
var prefix *string = flag.String("p", "iso19139", "The prefix")
var debug *bool = flag.Bool("d", false, "Enable debugging")


// Die with an error message
func die(msg string) {
    fmt.Fprintf(os.Stderr, "oaipmh: %s\n", msg)
    os.Exit(1)
}

func main() {
    ctx := &Context{}

    command.On("sets", "List sets", &SetsCommand{
        Ctx: ctx,
    })
    command.On("list", "List identifiers", &ListCommand{
        Ctx: ctx,
    })
    command.On("get", "Get records", &GetCommand{
        Ctx: ctx,
    })

    providerUrl := command.PreArg("provider", "URL to the OAI-PMH provider")

    // Parse the command
    command.Parse()

    // Create the OAI-PMH session
    ctx.Session = NewOaipmhSession(*providerUrl, *prefix)

    if (*debug) {
        ctx.Session.SetUrlTraceFunction(func(url string) {
            fmt.Fprintf(os.Stderr, ">> %s\n", url)
        })
    }

    // Run the command
    command.Run()
}
