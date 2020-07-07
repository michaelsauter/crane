test: imports
	@(go list ./... | grep -v "vendor/" | xargs -n1 go test -v -cover)

imports:
	@(goimports -w crane)

fmt:
	@(gofmt -w crane)

build: build-linux build-darwin build-darwin-pro build-windows build-windows-pro

build-linux: imports
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o crane_linux_amd64 -v github.com/michaelsauter/crane

build-darwin: imports
	GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 go build -o crane_darwin_amd64 -v github.com/michaelsauter/crane

build-windows: imports
	GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -o crane_windows_amd64.exe -v github.com/michaelsauter/crane
