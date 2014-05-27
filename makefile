GO = go
TARGET = oaipmh
VER = 1.1

TARGET_WIN32 = $(TARGET).exe

RELEASE_LINUX64 = oaipmh-$(VER)-linux-x86_64
RELEASE_WIN32 = oaipmh-$(VER)-win32

all: clean $(TARGET) test

deps:
	$(GO) get -u 'github.com/lmika/command'
	$(GO) get -u 'code.google.com/p/gcfg'
	$(GO) get -u 'github.com/nu7hatch/gouuid'
	$(GO) get -u 'launchpad.net/xmlpath'

clean:
	-rm $(TARGET)
	-rm $(TARGET).exe
	-rm -r $(RELEASE_LINUX64).zip
	-rm -r $(RELEASE_WIN32).zip

release: all $(RELEASE_LINUX64).zip $(RELEASE_WIN32).zip

$(RELEASE_LINUX64).zip: $(TARGET)
	mkdir $(RELEASE_LINUX64)
	cp $(TARGET) $(RELEASE_LINUX64)/.
	zip -r $(RELEASE_LINUX64).zip $(RELEASE_LINUX64)
	rm -r $(RELEASE_LINUX64)

$(RELEASE_WIN32).zip: $(TARGET_WIN32)
	mkdir $(RELEASE_WIN32)
	cp $(TARGET_WIN32) $(RELEASE_WIN32)/.
	zip -r $(RELEASE_WIN32).zip $(RELEASE_WIN32)
	rm -r $(RELEASE_WIN32)

$(TARGET): src/*.go
	( cd src ; $(GO) build -o ../$(TARGET) )

$(TARGET_WIN32): src/*.go
	( cd src ; GOOS=windows GOARCH=386 $(GO) build -o ../$(TARGET).exe )

test: oaipmh
	( cd src ; go test )
