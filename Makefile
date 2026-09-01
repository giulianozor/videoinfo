BINARY  := videoinfo
PREFIX  ?= /usr/local
GOFLAGS ?=

# cgo needs the FFmpeg development libraries. On macOS (Homebrew) they live
# in /opt/homebrew; override PKG_CONFIG_PATH if your layout differs.
export PKG_CONFIG_PATH ?= $(shell if [ -d /opt/homebrew/lib/pkgconfig ]; then printf /opt/homebrew/lib/pkgconfig; else printf /usr/local/lib/pkgconfig; fi)

.PHONY: all build test vet fmt clean install uninstall

all: build

build:
	go build $(GOFLAGS) -o $(BINARY) .

test:
	go test $(GOFLAGS) ./...

vet:
	go vet $(GOFLAGS) ./...

fmt:
	go fmt ./...

install: build
	install -d $(DESTDIR)$(PREFIX)/bin
	install -m 0755 $(BINARY) $(DESTDIR)$(PREFIX)/bin/$(BINARY)

uninstall:
	rm -f $(DESTDIR)$(PREFIX)/bin/$(BINARY)

clean:
	rm -f $(BINARY)
