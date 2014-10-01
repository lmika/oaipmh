package main

import (
    "log"
    "os"
    "os/user"
    "path/filepath"
    "strings"
    "errors"

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
func (cfg *Config) LookupProvider(endpoint string) (*Provider, error) {
    if endpoint == "" {
        return nil, nil
    }

    if prov, hasProv := cfg.Provider[endpoint] ; hasProv {
        return prov, nil
    } else if strings.HasPrefix(endpoint, "http") {
        return &Provider{ Url: endpoint }, nil
    } else {
        return nil, errors.New("Invalid endpoint or url: " + endpoint)
    }
}


func ReadConfig() *Config {
    c := &Config{
        Provider:  make(map[string]*Provider),
    }

    u, err := user.Current()
    if (err != nil) {
        log.Println("Error trying to get local user.  Using default config.  Error = %s\n", err.Error())
        return c
    }

    // Read the home config file
    homeConfig := filepath.Join(u.HomeDir, ".oaipmh.cfg")

    if _, err := os.Stat(homeConfig) ; err == nil {
        err := gcfg.ReadFileInto(c, homeConfig)
        if (err != nil) {
            panic(err)
        }
    }

    return c;
}
