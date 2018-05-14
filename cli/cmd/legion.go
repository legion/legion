package cmd

import (
	"fmt"
	"os"

	"github.com/legion/cli/backends"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	VERSION string
)

var cfgFile string
var backendFactory backends.BackendFactory

var RootCmd = &cobra.Command{
	Use:   "legion",
	Short: "legion CLI",
	Long:  `Run Serverless On-Demand Spark Jobs on Cloud Native platforms`,
}

func Execute(version string) {
	VERSION = version

	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	backendFactory = backends.NewBackendFactory()
	cobra.OnInitialize(initConfig)
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	// Find home directory.
	home, err := homedir.Dir()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// Search config in home directory with name ".cobra-example" (without extension).
	viper.AddConfigPath(home)
	viper.SetConfigName(".legion-conf")

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}
