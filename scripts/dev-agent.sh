#!/bin/bash
set -e

cd $(dirname $0)/../bin

# Prime sudo
sudo echo Compiling CLI
go build -tags k3s -o kube-api-auth-agent ../cli/main.go

echo Building image and agent
../image/build

echo Running
exec sudo ENTER_ROOT=../image/main.squashfs ./kube-api-auth-agent --debug agent -s https://localhost:7443 -t $(<${HOME}/.rancher/kube-api-auth/server/node-token)
