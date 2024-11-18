set GOOS=windows
set GOARCH=amd64
set CGO_ENABLED=0
go build -o token-gen.exe cmd\token-gen\main.go
go build -o forward.exe cmd\forward\main.go
set GOOS=linux
go build -o lambda cmd\lambda\main.go

./refresh.cmd
