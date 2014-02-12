package main

// Const
const DateFormat string = "2006-01-02"

type Context struct {
    Session         *OaipmhSession
    LogMsg          func (msg string)

    Config          *Config

    // Set to the provider is one is used instead of a raw URL
    Provider        *Provider
}
