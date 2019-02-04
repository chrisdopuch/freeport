// Copyright © 2019 Chris Dopuch <chris.dopuch@gmail.com>

package cmd

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strconv"
	"strings"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "freeport",
	Short: "A command line utility to free up a port",
	Long:  `Freeport is a CLI tool for ending any process listening on given port(s).`,
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			return errors.New("requires at least one argument")
		}
		for _, s := range args {
			if _, err := strconv.Atoi(s); err != nil {
				return fmt.Errorf("port arguments must be numbers, received '%s'", s)
			}
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Port(s):", strings.Join(args, " "))

		netstatCmd := exec.Command("netstat.exe", "-a", "-n", "-o")
		netstatOut, err := netstatCmd.Output()
		if err != nil {
			panic(err)
		}

		var grepArgs []string
		grepArgs = make([]string, len(args)*2)
		for i, s := range args {
			grepArgs[i] = fmt.Sprintf("-e :%s", s)
		}
		for i, j := 0, 0; i < len(args)*2; i, j = i+2, j+1 {
			grepArgs[i] = "-e"
			grepArgs[i+1] = fmt.Sprintf(":%s", args[j])
		}
		grepCmd := exec.Command("grep", grepArgs...)
		grepIn, inErr := grepCmd.StdinPipe()
		if inErr != nil {
			panic(inErr)
		}
		grepOut, outErr := grepCmd.StdoutPipe()
		if outErr != nil {
			panic(outErr)
		}
		grepCmd.Start()
		grepIn.Write(netstatOut)
		grepIn.Close()
		grepBytes, _ := ioutil.ReadAll(grepOut)
		grepCmd.Wait()

		fmt.Println(string(grepBytes))
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.freeport.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// Search config in home directory with name ".freeport" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigName(".freeport")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}
