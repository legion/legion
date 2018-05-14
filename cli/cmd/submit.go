package cmd

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/legion/cli/backends"

	"github.com/spf13/cobra"
)

var executorCores = 1
var executorMemory = "1.5g"
var numExecutors = 0
var driverMemory = "4g"
var driverCores = 2
var class string
var driverClassPath string
var existingMaster string
var keepAlive = false
var jars string
var backend string
var location string
var filePath string

var submitCmd = &cobra.Command{
	Use:   "submit",
	Short: "submit a spark job",
	Run: func(cmd *cobra.Command, args []string) {
		rand.Seed(time.Now().UnixNano())

		if existingMaster == "" && location == "" {
			fmt.Println("required flag(s) 'location' not set")
		}

		backendImpl := backendFactory.GetBackend(backend)

		if backendImpl == nil {
			fmt.Println("No backend named " + backend + " found")
		} else {
			fmt.Println(backend + " backend found")

			if existingMaster == "" && numExecutors == 0 {
				numExecutors = 2
			}

			response, err := backendImpl.Submit(executorCores, executorMemory, numExecutors, driverMemory, driverCores, location, existingMaster)

			if err != nil {
				fmt.Println(errors.New("failed submit command to backend  " + backend + ". error: " + err.Error()))
			}

			masterIP := response.MasterInfo.IPAddress
			params := getExtraParams()
			err = SubmitFileToMaster(response, filePath, params)

			if err != nil {
				fmt.Println("failed submitting file to remote master: " + err.Error())
			}

			fmt.Println("Spark job submitted. Master UI available at: " + "http://" + masterIP + ":8080")
		}
	},
}

func getExtraParams() map[string]string {
	params := map[string]string{}

	if class != "" {
		params["--class"] = class
	}

	if driverClassPath != "" {
		params["--driver-class-path"] = driverClassPath
	}

	if jars != "" {
		params["--jars"] = jars
	}

	return params
}

func SubmitFileToMaster(submitResponse backends.SubmitResponse, filePath string, params map[string]string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file", filepath.Base(filePath))

	if err != nil {
		return err
	}
	_, err = io.Copy(part, file)

	for key, val := range params {
		_ = writer.WriteField(key, val)
	}

	err = writer.Close()
	if err != nil {
		return err
	}

	url := "http://" + submitResponse.MasterInfo.IPAddress + ":7090/submit"

	resp, err := http.Post(url, writer.FormDataContentType(), body)
	if err != nil || resp.StatusCode != 200 {
		responseData, _ := ioutil.ReadAll(resp.Body)
		return errors.New(string(responseData))
	}

	if !keepAlive && existingMaster == "" {
		jsonString, _ := json.Marshal(submitResponse.WorkersInfo)
		url = "http://" + submitResponse.MasterInfo.IPAddress + ":7090/workers/expire"

		resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonString))
		if err != nil || resp.StatusCode != 200 {
			responseData, _ := ioutil.ReadAll(resp.Body)
			return errors.New(string(responseData))
		}
	}

	return nil
}

func init() {
	RootCmd.AddCommand(submitCmd)
	submitCmd.Flags().StringVar(&backend, "backend", "", "the serverless backend for spark (required)")
	submitCmd.Flags().IntVar(&executorCores, "executor-cores", executorCores, "cores for each executor")
	submitCmd.Flags().StringVar(&executorMemory, "executor-memory", executorMemory, "memory for each executor")
	submitCmd.Flags().IntVar(&numExecutors, "num-executors", numExecutors, "number of spark executors to lunch")
	submitCmd.Flags().StringVar(&driverMemory, "driver-memory", driverMemory, "memory for spark driver app")
	submitCmd.Flags().IntVar(&driverCores, "driver-cores", driverCores, "cores for spark driver app")
	submitCmd.Flags().StringVar(&class, "class", "", "the entry point for your application (e.g. org.apache.spark.examples.SparkPi)")
	submitCmd.Flags().StringVar(&filePath, "file", "", "path to jar or py file")
	submitCmd.Flags().StringVar(&location, "location", "", "region for provisioned resources")
	submitCmd.Flags().StringVar(&driverClassPath, "driver-class-path", "", "class paths for spark driver program")
	submitCmd.Flags().StringVar(&jars, "jars", "", "additional jars to be transffered to the cluster")
	submitCmd.Flags().StringVar(&existingMaster, "master", "", "existing Spark Master IP address or DNS name")
	submitCmd.Flags().BoolVar(&keepAlive, "keep-alive", keepAlive, "keep legion master and workers running after job is finished")
	submitCmd.MarkFlagRequired("file")
	submitCmd.MarkFlagRequired("backend")
}
