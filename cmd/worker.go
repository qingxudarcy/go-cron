/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"github.com/spf13/cobra"

	Init "go-cron/internal/worker/init"
)

// workerCmd represents the worker command
var workerCmd = &cobra.Command{
	Use:   "worker",
	Short: "Init go-cron worker",
	Long:  `Init go-cron worker service`,
	Run: func(cmd *cobra.Command, args []string) {
		var confFile string
		confFile, _ = cmd.Flags().GetString("config")
		Init.InitWorker(confFile)
	},
}

func init() {
	workerCmd.Flags().StringP("config", "c", "./config/worker.json", "指定配置文件路径")
	rootCmd.AddCommand(workerCmd)
}
