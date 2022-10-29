/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"os"

	"github.com/BurntSushi/toml"
	"github.com/spf13/cobra"
	"github.com/zzong12/hprotoxy/log"
	"github.com/zzong12/hprotoxy/server"
)

var configFile string

func init() {
	rootCmd.PersistentFlags().StringVarP(&configFile, "config", "C", "./config.toml", "config file path, default is ./config.toml")
}

var rootCmd = &cobra.Command{
	Use:   "hprotoxy",
	Short: "",
	Long:  ``,
	Args:  cobra.MinimumNArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		cfg := new(server.Config)
		if _, err := toml.DecodeFile(configFile, cfg); err != nil {
			log.Log.Fatalf("decode config file error: %v", err)
		}
		svr := server.NewServer(*cfg)
		svr.Run()
	},
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}
