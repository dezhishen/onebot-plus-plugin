#! /bin/bash
exclude=("plugins","pkg","script") 
for file in $(ls `dirname $0`/../) #注意此处这是两个反引号，表示运行系统命令
do
    if [ -d `dirname $0`/../$file ] #注意此处之间一定要加上空格，否则会报错
    then
        if [[ ! "${exclude[@]}" =~ "${file}" ]]; then
        echo "build plugin : $file"
            # echo "sh `dirname $0`/build_one.sh $file"
            echo "running : sh `dirname $0`/build_one.sh $file"
            sh `dirname $0`/build_one.sh $file
        fi 
    fi
done