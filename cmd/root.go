package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	c       string
	d       string
	rootCmd = &cobra.Command{
		Use:   "feedr",
		Short: "feedr is a standalone RSS feed aggregator",
		Long: `feedr's purpose for now is to generate content for the Sinhala blog aggregator. this could be a
				tool to be used by anyone to setup their own aggregator.`,
		Run: func(c *cobra.Command, a []string) {
			fmt.Fprintf(os.Stderr, "help text will be shown later here")
		},
	}
)

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "%s", err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&c, "config", "c", "", "configuration yaml file")
	rootCmd.MarkPersistentFlagRequired("config")
	//rootCmd.MarkPersistentFlagRequired("database")
}
