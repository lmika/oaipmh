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
var listProvidersFlag *bool = flag.Bool("P", false, "List providers and exit")


// Die with an error message
func die(msg string) {
    fmt.Fprintf(os.Stderr, "oaipmh: %s\n", msg)
    os.Exit(1)
}

// List the configured set of providers
func listProviders(ctx *Context) {
    for name, pconfig := range ctx.Config.Provider {
        fmt.Printf("%s = %s\n", name, pconfig.Url)
    }
}

func main() {
    ctx := &Context{
        Config: ReadConfig(),
        LogMsg: func(msg string) { },
    }

    command.On("sets", "List sets", &SetsCommand{
        Ctx: ctx,
    })
    command.On("list", "List identifiers", &ListCommand{
        Ctx: ctx,
    })
    command.On("get", "Get records", &GetCommand{
        Ctx: ctx,
    })
    command.On("harvest", "Harvest records and save them as files", &HarvestCommand{
        Ctx: ctx,
    })

    providerUrl := command.PreArg("provider", "URL to the OAI-PMH provider")

    // Parse the command
    res := command.TryParse()
    if (res != command.TryParseOK) {

        // Handle flags which do not require a command
        if (*listProvidersFlag) {
            listProviders(ctx)
            os.Exit(0)
        } else {
            command.Usage()
            os.Exit(1)
        }
    }

    // Create the OAI-PMH session
    ctx.Provider = ctx.Config.LookupProvider(*providerUrl)
    ctx.Session = NewOaipmhSession(ctx.Provider.Url, *prefix)

    if (*debug) {
        ctx.Session.SetUrlTraceFunction(func(url string) {
            fmt.Fprintf(os.Stderr, "[debug] %s\n", url)
        })
        ctx.LogMsg = func(msg string) {
            fmt.Fprintf(os.Stderr, "[debug] %s\n", msg)
        }
    }

    // Run the command
    command.Run()
}
