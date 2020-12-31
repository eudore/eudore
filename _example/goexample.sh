#!/bin/bash

rm -f *_test.go
for i in $(ls *.go)
do
	sed -i '1i\```go' $i
	echo '```' >> $i
	mv $i "${i:0:${#i} - 3}.md" 
done
