TARGET = oaipmh
VER = 1.0

RELEASE_ZIP = oaipmh-viewer-$(VER).zip
RELEASE_DIR = oaipmh-viewer-$(VER)

all: clean oaipmh

deps:
	go get 'github.com/moovweb/gokogiri'
	go get 'github.com/lmika/command'
	go get 'code.google.com/p/gcfg'

clean:
	-rm $(TARGET)
	-rm -r $(RELEASE_DIR)
	-rm -r $(RELEASE_ZIP)

release: clean all
	mkdir $(RELEASE_DIR)
	cp $(TARGET) $(RELEASE_DIR)
	zip -r $(RELEASE_ZIP) $(RELEASE_DIR)

oaipmh: src/*.go
	( cd src ; go build -o ../$(TARGET) )
