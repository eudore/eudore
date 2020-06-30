#!/bin/bash

# OUT=/mnt/hgfs/golang/coverage.html GOROOT=/usr/local/go1.11 bash gotest.sh

# 检测目录
dir=$(go list -json github.com/eudore/eudore  | grep '"Dir"' | cut -f4 -d'"')
if [ -z $dir ];then
	exit 1
fi
echo "goroot: $GOROOT"
$GOROOT/bin/go version
echo $dir

# 转换example
cd $dir/_example
for i in $(ls *.go | grep -v _test.go)
do
	funcname="$(echo ${i:0:1} | tr 'a-z' 'A-Z')${i:1:${#i} - 4}"
	testname=${i/\.go/_test.go}
	sed 's/func main()/func Test'"${funcname}"'(*testing.T)/' ${i} > $dir/${testname}
	sed -i 's/import/import "testing"\nimport/' $dir/${testname}
	sed -i 's/package main/package eudore_test/' $dir/${testname}
	sed -i 's/\/\/ app.CancelFunc()/app.CancelFunc()/' $dir/${testname}
done

# 复制文件
cp -rf *_test.go $dir/
cd $dir
rm -f appDefine_test.go appNotify_test.go appDaemon_test.go

COVERPKG='github.com/eudore/eudore,github.com/eudore/eudore/middleware'
export ENV_KEYS_NAME=eudore
export GODOC=https://golang.org

# 运行测试
if [ $# -ne 0 ];then
	$*
elif [ -z $OUT ];then
	$GOROOT/bin/go test -v -timeout=2m -cover -coverpkg=$COVERPKG $OPTION
else
	$GOROOT/bin/go test -v -timeout=2m -cover -coverpkg=$COVERPKG -coverprofile=size_coverage.out $OPTION && go tool cover -html=size_coverage.out -o $OUT && rm -f size_coverage.out
fi
rm -f *_test.go
