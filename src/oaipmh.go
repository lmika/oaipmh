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

const APP_NAME string = "oaipmh-viewer"
const APP_VERSION string = "0.1"

// Flags
var prefix *string = flag.String("p", "iso19139", "The record prefix to retrieve")
var debug *bool = flag.Bool("d", false, "Display debugging output")
var displayVersion *bool = flag.Bool("V", false, "Display version and exit")
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

// Display the version string
func displayVersionInfo() {
    fmt.Printf("%s %s\n", APP_NAME, APP_VERSION)
    os.Exit(0)
}

func main() {
    ctx := &Context{
        Config: ReadConfig(),
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
    command.On("search", "Harvest records and search the contents using XPath", &SearchCommand{
        Ctx: ctx,
    })
    command.On("serve", "Start a OAI-PMH provider to host the records on", &HostCommand{
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
        } else if (*displayVersion) {
            displayVersionInfo()
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
        ctx.Debug = true
        ctx.Session.SetDebug(true)
    }

    // Run the command
    command.Run()
}
