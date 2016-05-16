test:
	@(go list ./... | grep -v "vendor/" | xargs -n1 go test -v -cover)

build-linux-386:
	GOOS=linux GOARCH=386 CGO_ENABLED=0 go build -o crane_linux_386 -v github.com/michaelsauter/crane

build-linux-amd64:
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o crane_linux_amd64 -v github.com/michaelsauter/crane

build-darwin-amd64:
	GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 go build -o crane_darwin_amd64 -v github.com/michaelsauter/crane

build-windows-amd64:
	GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -o crane_windows_amd64.exe -v github.com/michaelsauter/crane
