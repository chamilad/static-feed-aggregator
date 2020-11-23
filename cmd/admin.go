package cmd

import (
	"fmt"
	"log"
	"os"

	"github.com/chamilad/sinhala-blog-aggregator/common"
	"github.com/spf13/cobra"
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
