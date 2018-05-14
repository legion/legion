package main

import (
	"fmt"

	"github.com/legion/cli/cmd"
)

func test_submit_python_file_upload() {
	err := cmd.SubmitFileToMaster("localhost", "mytestapp.py", map[string]string{})

	if err != nil {
		fmt.Println(err.Error())
	}
}

func main() {
	test_submit_python_file_upload()
}
