test:
	@(go list ./... | grep -v "vendor/" | xargs -n1 go test -v -cover)

build: build-linux-386 build-linux-amd64 build-darwin-amd64 build-windows-amd64

build-linux-386:
	GOOS=linux GOARCH=386 CGO_ENABLED=0 go build -o crane_linux_386 -v github.com/michaelsauter/crane

build-linux-amd64:
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o crane_linux_amd64 -v github.com/michaelsauter/crane

build-darwin-amd64:
	GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 go build -o crane_darwin_amd64 -v github.com/michaelsauter/crane

build-darwin-amd64-pro:
	GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 go build -tags pro -o crane_darwin_amd64_pro -v github.com/michaelsauter/crane

build-windows-amd64:
	GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -o crane_windows_amd64.exe -v github.com/michaelsauter/crane
