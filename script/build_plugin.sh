#! /bin/bash
exclude=("plugins","pkg","script") 
function read_dir(){
    for file in `ls $1` #注意此处这是两个反引号，表示运行系统命令
    do
        if [ -d $1"/"$file ] #注意此处之间一定要加上空格，否则会报错
        then
            echo "build plugin : $file"
            if [[ ! "${exclude[@]}" =~ "${file}" ]]; then
                export CGO_ENABLED=1
                export GOARCH=amd64
                export GOOS=windows
                go build -o ../plugins/$file-windows-amd64.exe ../$file
                export GOARM=7
                export GOARCH=arm64
                export GOOS=linux
                go build -o ../plugins/$file-linux-arm64v7 ../$file
                export GOARCH=amd64
                export GOOS=darwin
                go build -o ../plugins/$file-linux-amd64 ../$file
                export GOARCH=amd64
                export GOOS=linux
                go build -o ../plugins/$file-linux-amd64 ../$file
            fi 
        fi
    done
} 
#读取第一个参数
read_dir $1