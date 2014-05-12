GO = go
TARGET = oaipmh
VER = 1.1

RELEASE_ZIP = oaipmh-viewer-$(VER).zip
RELEASE_DIR = oaipmh-viewer-$(VER)

all: clean $(TARGET) test

deps:
	$(GO) get -u 'github.com/moovweb/gokogiri'
	$(GO) get -u 'github.com/lmika/command'
	$(GO) get -u 'code.google.com/p/gcfg'
	$(GO) get -u 'github.com/nu7hatch/gouuid'

clean:
	-rm $(TARGET)
	-rm -r $(RELEASE_DIR)
	-rm -r $(RELEASE_ZIP)

release: clean all
	mkdir $(RELEASE_DIR)
	cp $(TARGET) $(RELEASE_DIR)
	zip -r $(RELEASE_ZIP) $(RELEASE_DIR)

$(TARGET): src/*.go
	( cd src ; $(GO) build -o ../$(TARGET) )

test: oaipmh
	( cd src ; go test )
