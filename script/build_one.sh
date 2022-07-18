
#export CGO_ENABLED=1

export GOARCH=amd64
export GOOS=windows
go build -o ./$(dirname $0)/../plugins/$1-windows-amd64.exe ./$(dirname $0)/../$1

export GOARM=7
export GOARCH=arm64
export GOOS=linux

go build -o ./$(dirname $0)/../plugins/$1-linux-arm64v7 ./$(dirname $0)/../$1

export GOARCH=amd64
export GOOS=darwin
go build -o ./$(dirname $0)/../plugins/$1-linux-amd64 ./$(dirname $0)/../$1


export GOARCH=amd64
export GOOS=linux
go build -o ./$(dirname $0)/../plugins/$1-linux-amd64 ./$(dirname $0)/../$1
