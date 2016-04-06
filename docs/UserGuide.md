User Guide
==========


Usage
-----

    oaipmh [GLOBAL_FLAGS] PROVIDER COMMAND [ARGUMENTS]

Providers can either be a URL to a remote provider, or a provider alias (see Configuration below).  Command
is one of the following commands listed below.

### Global Flags

- `-d`: Enable debugging output.  Use `-dd` to increase the verbosity.
- `-p <prefix>`: Set the OAI-PMH prefix.  Default is "iso19139".
- `-P`: List the set of provider aliases, then exit.
- `-V`: Display the version number, then exit.
- `-G`: Use HTTP GET for requests instead of HTTP POST

Commands
--------

The tool supports the following commands.

### help

Show a brief description of each command.  Use `help <command>` to show usage details of a specific
command.

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
- `-l`: Display a detailed listing.  This will also display deleted identifiers, and will produce a summary of the number of metadata records to
    standard error.
- `-s`: Specify the set to retrieve.  Use '*' to list all sets.  When not specified, the "default set" is used if one is defined (see Configuration).
- `-d`: Show deleted records in the listing, along with active ones.
- `-D`: Only show deleted records.  Active ones will be hidden.
- `-R`: Use the ListRecords verb instead of ListIdentifiers verb.  This is useful mainly for testing.

By default only active identifiers are displayed.  To view deleted identifiers, use the `-l` flag.

**Example**: List all identifiers from WIS-GISC-MELBOURNE with date-stamps occurring after 2014-01-01

    $ oaipmh 'http://wis.bom.gov.au/openwis-user-portal/srv/en/oaipmh' list -s 'WIS-GISC-MELBOURNE' -A '2014-01-01'

### get

Retrieve records from the provider and display them to STDOUT.

    get [FLAGS] RECORD...

Supported flags are:

- `-H`: Display the header of the record instead of the record itself.
- `-S`: Specify the separator line to use when returning multiple records.
- `-p`: Run an external process with the metadata record.  See "External Processes" in the configuration section.
- `-t`: Test for the presence of records by getting them.  This will display the record identifiers with either a `+`
    indicating that the record was retrieved successfully, or a `-` if there was an error of some sort.

Following the flags is a list of identifiers to retrieve.  When multiple records are returned, they will be separated by a
separator, line with will either be the argument to `-S`, or 4 equal (`=`) signs.  Identifiers can be read from a file
by using `@filename`, which should be a text file with one identifier per line.  To read identifiers from STDIN, use `@-`.

When using `-p`, the external process will be invoked with each metadata record.  The output of the command will be
displayed to stdout.

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
- `-N <rs-expr>`: Evaluate the RS expression for each harvested record and use the result as the filename.  If the result of the RS Expression is *false*, the URN will be used (note: this may change in the future).  See *RS Expressions* below.
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

- `-A`, `-B`, `-c`, `-f`, `-s`: same as the flags of `list`.  These are used to select the records to search.

The query is an RS Expression which, when evaluated to true, will list the URN in the output.  For more information on RS Expressions,
see below.

Metadata that matches the search expression will be listed to stdout in the following form:

    <urn>: <searchResult>

**Example**: search for metadata records with an 'environmentDescription' element in all sets from the *eg* provider:

    $ oaipmh eg search -s "" 'xp("//environmentDescription")'

### compare

Compares the records from two providers.  The first provider is the provider that appears before the 'compare' command.

    compare <otherProvider>

Supported flags are:

- `-A`, `-B`, `-c`, `-f`, `-s`: same as the flags of `list`.  These are used to select the records to compare.
- `-F`: Read the identifiers to harvest from a file, instead of querying the OAI-PMH provider.  The file should be a text file with one identifier per line.
- `-C`: Compare the content of the metadata that appears in both providers.  This will increase the comparison time significantly.

The URN of differing records will be written as lines to stdout in the form `result urn`, where result is one of:

- `-`: URN exists in the first provider but is missing in the second provider.
- `+`: URN exists in the second provider but is missing from the first provider.
- `D`: URN exists in both providers but the contents differ (only applicable if `-C` is enabled)
- `E`: Fetching information about the URN has caused an error.

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


RS Expressions
--------------

*Record Search* expressions (or *RS Expressions*) is a very simple expression language which is evaluated over the contents
of a metadata record.  They are used mainly for the `search` command, but can also be found in other areas of the tool (e.g.
the `-N` option for the `harvest` command).

These expressions are similar to expressions found in any standard programming language.  A formal grammar of these expressions
are provided in BNF below:

    expression  :=  fncall | literal
    fncall      :=  IDENT "(" (expression ("," expression)*)? ")"
    literal     :=  STRING
    IDENT       :=  [a-zA-Z0-9]+
    STRING      :=  \" .* \"

At the moment, only strings are supported.  The expression will be considered *true* if the resulting string of the expression
is non-empty.

The functions supported by the language are:

Function | Description 
-------- | -----------
`concat(strs...)` | Returns a string which is all the individual arguments concatenated together.
`contains(str, substr)` | Returns *str* if it contains *substr*.  Otherwise, returns the empty string.
`replace(str, substr, newstr)` | Returns a string with all instances of *substr* within *str* replaced with *newstr*.
`startsWith(str, prefix)` | Returns *str* if it starts with *prefix*.  Otherwise, returns the empty string.
`urn()` | Returns the identifier of the record.
`xp(xpath)` | Performs an restricted XPath expression over the metadata and returns the element value that matches the path as the search result.  The XPath expression does not require name-spaces.


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

### External Processes

External processes can be used to configure common tools which consume metadata records.  These can be
used with commands like `get` which will run the metadata record through the tool, instead of displaying them.

    [extprocess "<name>"]
    cmd=<cmd>
    tempfile=<tempfile>

Configuration values to use:

- *name*: The name of the external process.
- *cmd*: The command to execute.
- *tempfile*: If "true", the record content will be written to a temporary file.  Otherwise, the content will be 
    piped to the command via stdin.

The command will be executed within the configured shell of the current user as determined by the SHELL
environment variable.  Note that the shell must support the `-c` switch in order to accept *cmd* as an
argument.  If *tempfile* is set to "false", the record content itself is piped to the command via
stdin.  Setting *tempfile* to true will have the record content written to a temporary file prior
to invoking the command.

Information about the current record is exposed to the subprocess through environment variables:

- *urn*: The record identifier
- *file*: The file containing the record content, if *tempfile* is set to true.

How the output of the command is used will depend on the command invoking the subprocess.  For example,
the `get` command will simply forward the output to stdout.  Anything written to stderr will always be
forward to stderr of `oaipmh`.  A command returning a non-zero return code will show an error and will
usually terminate processing of the command.
