#!/bin/bash
GOOS=linux GOARCH=arm go build
scp xmas-tree pi@192.168.219.173:~/
scp ballerine.gif pi@192.168.219.173:~/
