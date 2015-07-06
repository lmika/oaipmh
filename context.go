package main

// Const
const DateFormat string = "2006-01-02"

type LogLevel     int
const (
    NoLogLevel     LogLevel     = iota
    DebugLogLevel               = iota
    TraceLogLevel               = iota
)

type Context struct {
    // The logging level
    LogLevel        LogLevel

    Session         *OaipmhSession

    Config          *Config

    // Set to the provider is one is used instead of a raw URL
    Provider        *Provider
}
