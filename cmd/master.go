/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"github.com/spf13/cobra"

	Init "go-cron/internal/master/init"
)

// masterCmd represents the master command
var masterCmd = &cobra.Command{
	Use:   "master",
	Short: "Init master service",
	Long:  `Init go-cron api service`,
	Run: func(cmd *cobra.Command, args []string) {
		var confFile string
		confFile, _ = cmd.Flags().GetString("config")
		Init.InitMaster(confFile)
	},
}

func init() {
	masterCmd.Flags().StringP("config", "c", "./config/master.json", "指定配置文件路径")
	rootCmd.AddCommand(masterCmd)
}
