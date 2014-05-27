OAIPMH-VIEWER
=============

Command line tool for querying and retrieving records from OAI-PMI providers.

Installing
----------

To install the binary:

1. Download the release from [here](https://github.com/lmika-bom/oaipmh-viewer/releases/latest)
2. Unzip the archive
3. Add `oaipmh` to your PATH

To build from source:

1. Download the GO SDK from [golang.org](http://golang.org/)
2. Run `make deps all`
3. Add `oaipmh` to your PATH


Usage
-----

    oaipmh [GLOBAL_FLAGS] PROVIDER COMMAND [ARGUMENTS]

Providers can either be a URL to a remote provider, or a provider alias (see Configuration below).  Command
is one of the following commands listed below.

### Global Flags

- `-P`: List the set of provider aliases, then exit.
- `-d`: Enable debugging output.
- `-p <prefix>`: Set the OAI-PMH prefix.  Default is "iso19139".

Commands
--------

The tool supports the following commands.

### sets

Lists the sets published by the provider.

    sets [-l]

When used with the `-l` flag, displays a long listing of the set, which includes the set description.

### list

List the identifiers from a provider.

    list [FLAGS]

Supported flags are:

- `-A <yyyy-mm-dd>`: List identifiers with date-stamps occurring after the given date (in local time).
- `-B <yyyy-mm-dd>`: List identifiers with date-stamps occurring before the given date (in local time).
- `-c <count>`: Maximum number of results to return.  Defaults to 100,000.  Use -1 to return all results.
- `-f <number>`: The first result to return.  For example: `-f 3` will start listing from the 3rd result.
- `-l`: Display a detailed listing.  This will also display deleted identifiers.
- `-s`: Specify the set to retrieve.  When not specified, uses the "default set" if one is defined (see Configuration) or list identifiers from all sets.

By default only active identifiers are displayed.  To view deleted identifiers, use the `-l` flag.

**Example**: List all identifiers from WIS-GISC-MELBOURNE with date-stamps occurring after 2014-01-01

    $ oaipmh 'http://wis.bom.gov.au/openwis-user-portal/srv/en/oaipmh' list -s 'WIS-GISC-MELBOURNE' -A '2014-01-01'

### get

Retrieve records from the provider and display them to STDOUT.

    get [FLAGS] RECORD...

Supported flags are:

- `-H`: Display the header of the record instead of the record itself.
- `-R`: Display the entire OAI-PMH response.  Useful for debugging.
- `-S`: Specify the separator line to use when returning multiple records.

Following the flags is a list of identifiers to retrieve.  When multiple records are returned, they will be separated by a
separator, line with will either be the argument to `-S`, or 4 equal (`=`) signs.  Identifiers can be read from a file
by using `@filename`, which should be a text file with one identifier per line.  To read identifiers from STDIN, use `@-`.

### harvest

Retrieve records from a provider and save them as files.  This command combines the use of `list` and `get`, while
providing some useful utilities for managing the saved files.

    harvest [FLAGS]

Supported flags are:

- `-A`, `-B`, `-c`, `-f`, `-s`: same as the flags of `list`.  These are used to select the records to retrieve.
- `-C`: Zip directories of harvested files once full.  Requires `zip` to be in the path.
- `-D <count>`: Maximum number of files to store in each directory.  Defaults to 10000.
- `-F`: Read the identifiers to harvest from a file, instead of querying the OAI-PMH provider.  The file should be a text file with one identifier per line.  Implies `-L`.
- `-L`: Retrieve records using separate GetRecord HTTP requests for each identifier.  Slower, but is less prone to errors when harvesting a large number of records.
- `-W`: Set the number of threads used to download records.  Only applicable when used with either `-L` or `-F`.
- `-n`: Dry run.  Do not save any records.

Records are stored in directories of the form *timestamp*/*subdirNo* where *timestamp* is the time the harvesting task was
started, and *subdirNo* is a monotonically increasing number.  Records are stored with the filename *identifier*.xml.

**Example**: harvest all records from WIS-GISC-MEBOURNE with date-stamps occurring after 2014-01-01

    $ oaipmh 'http://wis.bom.gov.au/openwis-user-portal/srv/en/oaipmh' harvest -s 'WIS-GISC-MELBOURNE' -A '2014-01-01'

**Example**: harvest records with identifiers found in *urns.txt* from GISC Exeter using 8 download threads, 20000 files per
directory and compressing directories once filled:

    $ oaipmh 'http://wis.metoffice.gov.uk/openwis-user-portal/srv/oaipmh' harvest -F urns.txt -D 20000 -W 8 -C

### search

Retrieve records from a provider and performs a search query over them.

    search [FLAGS] QUERY

Supported flags are:

- `-A`, `-B`, `-c`, `-f`, `-s`: same as the flags of `list`.  These are used to select the records to retrieve.
- `-F`, `-L`: same as the flags of `harvest`.  *Not yet supported*

The query is an expression with the following syntax:

    query       :=  predicate
    predicate   :=  predicatename "(" argument ")"

Valid predicates are:

- *xpath*: `xp( <xpath> )` - Performs an restricted XPath expression over the metadata and returns the element value that matches the path as the search result.  The XPath expression does not require namespaces.

Metadata that matches the search expression will be listed to stdout in the following form:

    <urn>: <searchResult>

**Example**: search for metadata records with an 'environmentDescription' element in all sets from the *eg* provider:

    $ oaipmh eg search -s "" 'xp("//environmentDescription")'

### serve

Starts a temporary OAI-PMH endpoint and serves metadata organised into files and directories.  Used mainly for testing.

    serve 

The tool is to be started in the directory containing the files to serve.
The provider URL is treated as the hostname and port that the endpoint will listen on.

The tool expects all metadata to be arranged into directories, with each directory representing a set.  The directory name
will be used as the set name and the metadata within the directory will belong to that set.  Records must be XML: non XML
files will not be recognised by the endpoint.  Record files must 

**Example**: start serving all metadata managed in the current directory over port 8080 on localhost.

    $ oaipmh "localhost:8080" serve 

**Example**: session showing arrangement of files and directories

    $ tree
        set1
            record1.xml
            record2.xml
        set2
            record3.xml
            record4.xml
    $ oaipmh "localhost:8080" serve &
    $ oaipmh "http://localhost:8080/" sets
    set1
    set2
    $ oaipmh "http://localhost:8080/" list -s set1
    record1
    record2



Configuration
-------------

When starting, the oaipmh tool will look for configuration options at `~/.oaipmh.cfg`.

### Provider Aliases

Provider aliases are a way to define configuration options for commonly used providers.  Once defined, the
provider name can be used as the provider argument to oaipmh, instead of typing in the full provider URL.
They can also be used to setup the context for commands like `list` and `harvest` by, for example, setting
the default set.

    [provider "<name>"]
    url=<url>
    set=<defaultSet>

Configuration values to use:

- *name*: The provider alias name.  Can be anything, although URLs are discouraged.
- *url*: The URL of the OAI-PMH provider.
- *set*: The default set to use.  When `-s` is not specified in commands that use it (like `list` or `harvest`), this
set will be used instead.
