all: linux darwin windows
	
linux:
	CGO_ENABLED=0 GOOS=linux go build -o bin/mongostatusd_linux ./cmd/mongostatusd/
darwin:
	CGO_ENABLED=0 GOOS=darwin go build -o bin/mongostatusd_darwin ./cmd/mongostatusd/
windows:
	CGO_ENABLED=0 GOOS=windows go build -o bin/mongostatusd_windows ./cmd/mongostatusd/
