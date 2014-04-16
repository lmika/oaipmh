package main


import (
    "./oaipmh"

    "flag"
    "log"
    "net/http"
)


// ---------------------------------------------------------------------------------------------------
// Host command
//      Starts a HTTP OAI-PMH endpoint

type HostCommand struct {
    Ctx             *Context
}

func (gc *HostCommand) Flags(fs *flag.FlagSet) *flag.FlagSet {
    return fs
}

func (gc *HostCommand) Run(args []string) {
    bindUrl := gc.Ctx.Provider.Url

    repo := oaipmh.NewFileRepository(".")
    handler := oaipmh.NewHandler(repo)

    server := &http.Server{
        Addr: bindUrl,
        Handler: handler,
    }

    log.Printf("OAI-PMH provider running at %s", bindUrl)
    server.ListenAndServe()
}
