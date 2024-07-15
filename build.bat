set GOOS=linux
set GOARCH=arm64
go build -o "build/play arm64"

set GOOS=linux
set GOARCH=amd64
go build -o "build/play amd64"

set GOOS=
set GOARCH=
go build -o "build/play windows"
