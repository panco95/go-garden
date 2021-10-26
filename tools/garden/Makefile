build:
	go build -ldflags="-s -w" garden.go
	$(if $(shell command -v upx), upx gctl)
mac:
	GOOS=darwin go build -ldflags="-s -w" -o gctl-darwin garden.go
	$(if $(shell command -v upx), upx gctl-darwin)
win:
	GOOS=windows go build -ldflags="-s -w" -o gctl.exe garden.go
	$(if $(shell command -v upx), upx gctl.exe)
linux:
	GOOS=linux go build -ldflags="-s -w" -o gctl-linux garden.go
	$(if $(shell command -v upx), upx gctl-linux)