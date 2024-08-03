set GOOS=windows
set GOARCH=amd64
set CGO_ENABLED=0
go build -o token-gen.exe cmd\token-gen\main.go
go build -o forward.exe cmd\forward\main.go
set GOOS=linux
go build -o lambda cmd\lambda\main.go

set AWS_REGION=us-west-2
set AWS_ACCESS_KEY=$TOKENGEN_AWS_ACCESS_KEY
set AWS_SECRET_KEY=$TOKENGEN_AWS_SECRET_KEY
set LAMBDA_BUCKET=email-forward-lambda-package
token-gen.exe
