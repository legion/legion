#!/bin/bash

mkdir -p dist/linux
mkdir -p dist/darwin
mkdir -p dist/windows

gox -osarch="linux/amd64" -output "dist/linux/legion"
gox -osarch="darwin/amd64" -output "dist/darwin/legion"
gox -osarch="windows/amd64" -output "dist/windows/legion"