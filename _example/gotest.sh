#!/bin/bash

# 将example main转换成go test文件
mkdir -p test
rm -f test/*_test.go
for i in $(ls *.go | grep -v _test.go)
do
	funcname="$(echo ${i:0:1} | tr 'a-z' 'A-Z')${i:1:${#i} - 4}"
	testname=${i/\.go/_test.go}
	sed 's/func main()/func Test'"${funcname}"'(*testing.T)/' ${i} > test/${testname}
	sed -i 's/import/import "testing"\nimport/' test/${testname}
done

rm -f test/appCoreNotify_test.go
rm -f test/appEudoreDaemon_test.go
rm -f test/appEudoreNotify_test.go

go test -v test/*_test.go
rm -rf test

