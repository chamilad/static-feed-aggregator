package cmd

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/chamilad/sinhala-blog-aggregator/common"
	"github.com/spf13/cobra"
)

const (
	FEEDTABLE = "feeds"
)

var (
	adminCmd = &cobra.Command{
		Use:   "admin",
		Short: "a tool to manage the config yaml",
		Long:  "",
		Run:   Admin,
	}

	adminResetDatesCmd = &cobra.Command{
		Use:   "reset",
		Short: "reset last read timestamps and read all available feeds",
		Long:  "",
		Run:   AdminResetDates,
	}

	title   string
	url     string
	feedUrl string
	email   string
)

func init() {
	rootCmd.AddCommand(adminCmd)
	adminCmd.AddCommand(adminResetDatesCmd)
}

func Admin(cmd *cobra.Command, a []string) {
	fmt.Fprintf(os.Stderr, "a tool to manage the config.yaml")
}

func AdminResetDates(cmd *cobra.Command, a []string) {
	conf, err := common.LoadConfig(c)
	if err != nil {
		log.Fatalf("error while loading configuration: %s", err)
	}

	for i, fC := range conf.Aggr.Collector.Feeds {
		fC.LastRead = 0
		conf.Aggr.Collector.Feeds[i] = fC
	}

	err = common.UpdateConfig(conf, c)
	if err != nil {
		log.Fatalf("error while updating configuration: %s", err)
	}

	fmt.Fprintf(os.Stderr, "last read timestamps reset")
}

func initFeedTable(t string, d *sql.DB) error {
	s, err := d.Prepare(fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s (id INTEGER PRIMARY KEY AUTOINCREMENT, title TEXT UNIQUE, url TEXT UNIQUE, feed TEXT UNIQUE, email TEXT, added INTEGER, last_read INTEGER)", t))

	if err != nil {
		log.Fatalf("error while prepping create table: %s\n", err)
	}

	defer s.Close()

	_, err = s.Exec()
	if err != nil {
		log.Fatalf("error while executing create table: %s\n", err)
	}

	return nil
}
