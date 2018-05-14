package main

import "github.com/legion/cli/cmd"

var (
	VERSION = "0.0.1"
)

func main() {
	cmd.Execute(VERSION)
}
