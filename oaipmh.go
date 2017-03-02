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

	"sort"

	"github.com/lmika/command"
)

const APP_NAME string = "oaipmh"
const APP_VERSION string = "1.4"

// Flags
var prefix *string = flag.String("p", "iso19139", "The record prefix to retrieve")
var debug *bool = flag.Bool("d", false, "Show debug messages")
var debugLots *bool = flag.Bool("dd", false, "Show trace messages")
var displayVersion *bool = flag.Bool("V", false, "Display version and exit")
var listProvidersFlag *bool = flag.Bool("P", false, "List providers and exit")
var useGetFlag *bool = flag.Bool("G", false, "Use HTTP GET instead of HTTP POST")

// Die with an error message
func die(msg string) {
	fmt.Fprintf(os.Stderr, "oaipmh: %s\n", msg)
	os.Exit(1)
}

// List the configured set of providers
func listProviders(ctx *Context) {
	// Sort the names
	sortedNames := make([]string, 0, len(ctx.Config.Provider))
	for name := range ctx.Config.Provider {
		sortedNames = append(sortedNames, name)
	}

	sort.Strings(sortedNames)

	for _, name := range sortedNames {
		pconfig := ctx.Config.Provider[name]
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

	command.OnHelpShowUsage()
	command.OnHelpIgnorePreargs()

	command.On("compare", "Compare providers", &CompareCommand{Ctx: ctx}).Arguments("otherProvider")
	command.On("sets", "List sets", &SetsCommand{Ctx: ctx}).Arguments()
	command.On("list", "List identifiers", &ListCommand{Ctx: ctx}).Arguments()
	command.On("get", "Get records", &GetCommand{Ctx: ctx}).Arguments("record", "...")
	command.On("harvest", "Harvest records and save them as files", &HarvestCommand{Ctx: ctx}).Arguments()
	command.On("search", "Harvest records and search the contents using XPath", &SearchCommand{Ctx: ctx}).Arguments("expr")
	command.On("serve", "Start a OAI-PMH provider to host the records on", &HostCommand{Ctx: ctx}).Arguments()

	providerUrl := command.PreArg("provider", "URL to the OAI-PMH provider")

	// Parse the command
	err := command.TryParse()
	if err != nil {

		// Handle flags which do not require a command.
		if *listProvidersFlag {
			listProviders(ctx)
			os.Exit(0)
		} else if *displayVersion {
			displayVersionInfo()
			os.Exit(0)
		} else {
			err.(command.TryParseError).Usage()
			os.Exit(1)
		}
	}

	// If the provider Url is "help", display the usage
	if *providerUrl == "" {
		command.Usage()
		os.Exit(1)
	}

	// Create the OAI-PMH session
	ctx.Provider = ctx.Config.LookupProvider(*providerUrl)
	if ctx.Provider != nil {
		ctx.Session = NewOaipmhSession(ctx.Provider.Url, *prefix)
	}

	debugLevel := 0
	if *debugLots {
		debugLevel = 2
	} else if *debug {
		debugLevel = 1
	}

	if (debugLevel >= 0) && (ctx.Session != nil) {
		ctx.LogLevel = LogLevel(debugLevel)
		ctx.Session.SetDebug(debugLevel)
	}
	ctx.Session.SetUseGet(*useGetFlag)

	// Run the command
	command.Run()
}
