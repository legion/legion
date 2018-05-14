package cmd

import (
	"errors"
	"fmt"
	"math/rand"
	"time"

	"github.com/spf13/cobra"
)

var masterMemory = "4g"
var masterCores = 2
var masterBackend string
var masterLocation string

var deployCmd = &cobra.Command{
	Use:   "deploy-master",
	Short: "deploy a legion master",
	Run: func(cmd *cobra.Command, args []string) {
		rand.Seed(time.Now().UnixNano())

		backendImpl := backendFactory.GetBackend(masterBackend)

		if backendImpl == nil {
			fmt.Println("No backend named " + masterBackend + " found")
		} else {
			fmt.Println(backend + " backend found")
			master, err := backendImpl.CreateMaster(masterCores, masterMemory, masterLocation)

			if err != nil {
				fmt.Println(errors.New("failed creating master  " + masterBackend + ". error: " + err.Error()))
			}

			fmt.Println("Master created. FQDN: " + master.FQDN + ". IP: " + master.IPAddress + ". UI available at: " + "http://" + master.IPAddress + ":8080")
		}
	},
}

func init() {
	RootCmd.AddCommand(deployCmd)
	deployCmd.Flags().StringVar(&masterBackend, "backend", "", "the serverless backend for spark (required)")
	deployCmd.Flags().StringVar(&masterMemory, "driver-memory", masterMemory, "memory for the spark master and driver")
	deployCmd.Flags().IntVar(&masterCores, "driver-cores", masterCores, "cores for spark master and driver")
	deployCmd.Flags().StringVar(&masterLocation, "location", "", "region for provisioned resources")
	deployCmd.MarkFlagRequired("backend")
	deployCmd.MarkFlagRequired("location")
}
