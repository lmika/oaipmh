package main

import (
    "os"

    "code.google.com/p/gcfg"
)


// The provider struct
type Provider struct {
    // The provider URL
    Url         string

    // The default set
    Set         string
}

// Baseline configuration
type Config struct {
    // Provider aliases
    Provider      map[string]*Provider
}

// Looks up a provider.  If one is not defined, creates a dummy provider.
func (cfg *Config) LookupProvider(url string) *Provider {
    if prov, hasProv := cfg.Provider[url] ; hasProv {
        return prov
    } else {
        return &Provider{
            Url:    url,
        }
    }
}


func ReadConfig() *Config {
    c := &Config{
        Provider:  make(map[string]*Provider),
    }

    // Read the home config file
    homeConfig := os.ExpandEnv("${HOME}/.oaipmh.cfg")

    if _, err := os.Stat(homeConfig) ; err == nil {
        err := gcfg.ReadFileInto(c, homeConfig)
        if (err != nil) {
            panic(err)
        }
    }

    return c;
}
