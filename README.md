OAIPMH-VIEWER
=============

Command line tool for querying and retrieving records from OAI-PMI providers.

Installing
----------

Download the latest release [here](releases/latest).

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
