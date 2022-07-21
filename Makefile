test: imports
	@(go list ./... | grep -v "vendor/" | xargs -n1 go test -v -cover)

imports:
	@(goimports -w crane)

fmt:
	@(gofmt -w crane)

build: build-linux build-linux-arm build-linux-arm64 build-darwin build-darwin-arm64 build-windows

build-linux: imports
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o crane_linux_amd64 -v github.com/michaelsauter/crane/v3

build-linux-arm: imports
	GOOS=linux GOARCH=arm CGO_ENABLED=0 go build -o crane_linux_arm -v github.com/michaelsauter/crane/v3

build-linux-arm64: imports
	GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -o crane_linux_arm64 -v github.com/michaelsauter/crane/v3

build-darwin: imports
	GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 go build -o crane_darwin_amd64 -v github.com/michaelsauter/crane/v3

build-darwin-arm64: imports
	GOOS=darwin GOARCH=arm64 CGO_ENABLED=0 go build -o crane_darwin_arm64 -v github.com/michaelsauter/crane/v3

build-windows: imports
	GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -o crane_windows_amd64.exe -v github.com/michaelsauter/crane/v3
