test:
	@(go list ./... | grep -v "vendor/" | xargs -n1 go test -v)
