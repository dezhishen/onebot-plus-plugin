
export CGO_ENABLED=0

export GOARCH=amd64
export GOOS=windows
go build -o ../plugins/baidu_translate-windows-amd64.exe ../baidu_translate
go build -o ../plugins/cat_and_dog-windows-amd64.exe ../cat_and_dog
go build -o ../plugins/random-windows-amd64.exe ../random

export GOARM=7
export GOARCH=arm64
export GOOS=linux

go build -o ../plugins/baidu_translate-linux-arm64v7 ../baidu_translate
go build -o ../plugins/cat_and_dog-linux-arm64v7 ../cat_and_dog
go build -o ../plugins/random-linux-arm64v7 ../random

export GOARCH=amd64
export GOOS=darwin
go build -o ../plugins/baidu_translate-linux-amd64 ../baidu_translate
go build -o ../plugins/cat_and_dog-linux-amd64 ../cat_and_dog
go build -o ../plugins/random-linux-amd64 ../random


export GOARCH=amd64
export GOOS=linux
go build -o ../plugins/baidu_translate-linux-amd64 ../baidu_translate
go build -o ../plugins/cat_and_dog-linux-amd64 ../cat_and_dog
go build -o ../plugins/random-linux-amd64 ../random
