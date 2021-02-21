package cmd

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/chamilad/sinhala-blog-aggregator/common"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

var (
	adminListFeedsCmd = &cobra.Command{
		Use:   "listFeeds",
		Short: "list the feeds that are being aggregated by this instance",
		Long:  "",
		Run:   AdminListFeeds,
	}
)

func init() {
	adminCmd.AddCommand(adminListFeedsCmd)
}

func AdminListFeeds(cmd *cobra.Command, a []string) {
	fmt.Fprintf(os.Stderr, "looking up feeds...\n")

	conf, err := common.LoadConfig(c)
	if err != nil {
		log.Fatalf("error while loading configuration file: %s", err)
	}

	db, err := common.OpenDb(conf.Database)
	if err != nil {
		log.Fatalf("error while opening database: %s\n", err)
	}
	defer db.Close()

	initFeedTable(FEEDTABLE, db)

	feeds, err := db.Query(fmt.Sprintf("SELECT id, title, url, feed, email, added, last_read FROM %s ORDER BY added DESC", FEEDTABLE))
	if err != nil {
		log.Fatalf("error while retrieving feeds: %s", err)
	}

	w := tablewriter.NewWriter(os.Stdout)
	w.SetHeader([]string{"#", "Title", "URL", "Feed URL", "Contact Email", "Added On", "Last Read"})
	rc := 0
	for feeds.Next() {
		rc += 1
		var id int64
		var title string
		var url string
		var feed string
		var email string
		var added int64
		var lastRead int64

		err = feeds.Scan(&id, &title, &url, &feed, &email, &added, &lastRead)
		if err != nil {
			log.Fatalf("error while parsing feed entry: %s", err)
		}

		w.Append([]string{string(id), title, url, feed, email, time.Unix(added, 0).Format("2006-01-03 15:04:06"), time.Unix(lastRead, 0).Format("2006-01-03 15:04:06")})
	}

	if rc > 0 {
		w.Render()
	}

	fmt.Fprintf(os.Stderr, "all done")
}
