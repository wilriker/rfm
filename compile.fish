#!/usr/bin/env fish

set -l gofile cmd/rfm/main.go
set -l out rfm

env GOOS=linux GOARCH=arm go build -tags netgo -o $out $gofile
and tar czf $out-linux_arm.tgz $out LICENSE

env GOOS=linux GOARCH=arm64 go build -tags netgo -o $out $gofile
and tar czf $out-linux_arm64.tgz $out LICENSE

env GOOS=windows go build -tags netgo -o $out.exe $gofile
and zip -r $out-windows_amd64.zip $out.exe LICENSE

env GOOS=darwin go build -tags netgo -o $out $gofile
and tar czf $out-darwin_amd64.tgz $out LICENSE

env GOOS=linux go build -tags netgo -o $out $gofile
and tar czf $out-linux_amd64.tgz $out LICENSE
