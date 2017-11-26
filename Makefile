.PHONY: build dep clean fmt test

SCRAPER := cmd/wp-scraper

build: clean $(SCRAPER)

dep:
	dep ensure

clean:
	-rm -f $(SCRAPER)

fmt:
	go fmt ./...

test:
	go test

%: %.go
	go build -o $@ $<