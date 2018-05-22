#!/bin/bash

tag=$1
tmpgopath=`mktemp -d`
mkdir -p $tmpgopath/src/github.com/heetch
cp -r `pwd` $tmpgopath/src/github.com/heetch/.

GOPATH=$tmpgopath godoc -goroot $tmpgopath -http :6060 &
pid=$!

sleep 5

goscrape http://localhost:6060/pkg && kill $pid
mkdir -p doc/
mv "localhost:6060" doc/$tag
