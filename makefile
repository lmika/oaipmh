all: clean oaipmh

clean:
	-rm oaipmh

oaipmh: src/*.go
	( cd src ; go build -o ../oaipmh )
