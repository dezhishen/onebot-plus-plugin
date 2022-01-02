
export CGO_ENABLED=0

export GOARCH=amd64
export GOOS=windows
go build -o ../plugins/$1-windows-amd64.exe ../$1

export GOARM=7
export GOARCH=arm64
export GOOS=linux

go build -o ../plugins/$1-linux-arm64v7 ../$1

export GOARCH=amd64
export GOOS=darwin
go build -o ../plugins/$1-linux-amd64 ../$1


export GOARCH=amd64
export GOOS=linux
go build -o ../plugins/$1-linux-amd64 ../$1
