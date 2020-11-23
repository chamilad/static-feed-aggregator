package cmd

import (
	"fmt"
	"os"

	_ "github.com/mattn/go-sqlite3"
	"github.com/spf13/cobra"
)

var (
	cleanCmd = &cobra.Command{
		Use:   "cleanup",
		Short: "cleanup older items from the database",
		Long:  "",
		Run:   Cleanup,
	}
)

func init() {
	rootCmd.AddCommand(cleanCmd)
}

func Cleanup(cmd *cobra.Command, a []string) {
	fmt.Fprintf(os.Stderr, "%s not implemented yet", cmd.Name)
	// to be run once a month on the month start
	// 1. get all the posts from last month
	// 2. write an archive yaml file to data/archives/month-year.md
	// 3. ???
	// 4. profit
}
