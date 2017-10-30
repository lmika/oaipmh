OAIPMH-VIEWER
=============

Command line tool for querying and retrieving records from OAI-PMI providers.

Installing
----------

If you have [go](http://golang.org/):

1. Run `go get github.com/lmika/oaipmh`

To install the binary:

1. Download the release from [here](https://github.com/lmika/oaipmh/releases/latest)
2. Unzip the archive
3. Add `oaipmh` to your PATH

To build from source:

1. Download the GO SDK from [golang.org](http://golang.org/)
2. Checkout the source
3. Run `go install ./...`

Basic Usage
-----------

    oaipmh PROVIDER COMMAND

Where *provider* is a URL to an OAI-PMH endpoint and *command* is one of:

- [compare](docs/UserGuide.md#compare): Compare providers
- [get](docs/UserGuide.md#get): Get records
- [harvest](docs/UserGuide.md#harvest): Harvest records and save them as files
- [help](docs/UserGuide.md#help): Displays usage string of commands
- [list](docs/UserGuide.md#list): List identifiers
- [search](docs/UserGuide.md#search): Harvest records and search the contents using XPath
- [serve](docs/UserGuide.md#serve): Start a OAI-PMH provider to host the records on
- [sets](docs/UserGuide.md#sets): List sets

See the [User Guide](docs/UserGuide.md) for more details.

Licence
-------

Copyright (c) 2017 Leon Mika (Australian Bureau of Meteorology)

Permission is hereby granted, free of charge, to any person obtaining a copy of this software and
associated documentation files (the "Software"), to deal in the Software without restriction, including
without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is furnished to do so, subject to the
following conditions:

The above copyright notice and this permission notice shall be included in all copies or substantial
portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT
LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. 
IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY,
WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE
SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.