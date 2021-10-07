build:
	go build -ldflags="-s -w" gctl.go
	$(if $(shell command -v upx), upx gctl)
mac:
	GOOS=darwin go build -ldflags="-s -w" -o gctl-darwin gctl.go
	$(if $(shell command -v upx), upx gctl-darwin)
win:
	GOOS=windows go build -ldflags="-s -w" -o gctl.exe gctl.go
	$(if $(shell command -v upx), upx gctl.exe)
linux:
	GOOS=linux go build -ldflags="-s -w" -o gctl-linux gctl.go
	$(if $(shell command -v upx), upx gctl-linux)