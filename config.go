package main

import (
    "errors"
    "log"
    "os"
    "fmt"
    "io"
    "crypto/md5"
    "os/user"
    "os/exec"
    "path/filepath"
    "strings"
    "time"

    "gopkg.in/gcfg.v1"

    "github.com/lmika-bom/oaipmh/client"
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

    // External processes
    ExtProcess    map[string]*ExtProcess
}

// Looks up a provider.  If one is not defined, creates a dummy provider.
func (cfg *Config) LookupProvider(endpoint string) *Provider {
    if endpoint == "" {
        return nil
    }

    if prov, hasProv := cfg.Provider[endpoint] ; hasProv {
        return prov
    } else {
        return &Provider{ Url: endpoint }
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


// The external process configuration
type ExtProcess struct {
    // The shell command to execute
    Cmd         string

    // When true, the metadata will be written to a temp file and the filename will be
    // provided as a shell parameter "file"
    TempFile    bool
}

// Invoke this external process configuration with the given Oaipmh record
func (ep *ExtProcess) invokeWithRecord(rec *oaipmh.OaipmhRecord) error {
    shell, hasShell := os.LookupEnv("SHELL")
    if !hasShell {
        return errors.New("No SHELL defined")
    }

    // Setup the command
    cmd := exec.Command(shell, "-c", ep.Cmd)
    
    cmd.Env = os.Environ()
    cmd.Env = append(cmd.Env, "urn=" + rec.Header.Identifier)

    // Setup the metadata content
    if ep.TempFile {
        tmpFileName, err := ep.writeToTempFile(rec)
        if err != nil {
            return err
        }
        defer os.Remove(tmpFileName)

        cmd.Env = append(cmd.Env, "file=" + tmpFileName)
        cmd.Stdin = nil
    } else {
        cmd.Stdin = strings.NewReader(rec.Content.Xml)
    }

    cmd.Stdout = os.Stdout
    cmd.Stderr = os.Stderr

    return cmd.Run()
}

func (ep *ExtProcess) writeToTempFile(rec *oaipmh.OaipmhRecord) (string, error) {
    tmpFilename := fmt.Sprintf("oaipmh-%d-%x.xml", time.Now().UnixNano(), md5.Sum([]byte(rec.Header.Identifier)))

    fullPath := filepath.Join(os.TempDir(), tmpFilename)

    file, err := os.Create(fullPath)
    if err != nil {
        return "", err
    }
    defer file.Close()

    _, err = io.Copy(file, strings.NewReader(rec.Content.Xml))
    return fullPath, err
}