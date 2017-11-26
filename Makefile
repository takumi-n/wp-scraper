.PHONY: build dep clean fmt

SCRAPER := cmd/wp-scraper

build: clean $(SCRAPER)

dep:
	dep ensure

clean:
	-rm -f $(SCRAPER)

fmt:
	go fmt ./...

%: %.go
	go build -o $@ $<