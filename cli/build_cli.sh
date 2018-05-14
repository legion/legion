#!/bin/bash

VERSION="0.0.1"

mkdir -p dist/linux
mkdir -p dist/darwin
mkdir -p dist/windows

gox -osarch="linux/amd64" -output "dist/linux/legion"
gox -osarch="darwin/amd64" -output "dist/darwin/legion"
gox -osarch="windows/amd64" -output "dist/windows/legion"

tar -cvf dist/linux/legion_"$VERSION"_linux_amd64.tar dist/linux/legion
tar -cvf dist/darwin/legion_"$VERSION"_darwin_amd64.tar dist/darwin/legion
tar -cvf dist/windows/legion_"$VERSION"_windows_amd64.tar dist/windows/legion.exe