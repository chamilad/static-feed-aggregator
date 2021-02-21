package cmd

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/chamilad/sinhala-blog-aggregator/common"
	"github.com/spf13/cobra"
)

var (
	adminRemoveFeedCmd = &cobra.Command{
		Use:   "removeFeed",
		Short: "remove a feed from being aggregated",
		Long:  "",
		Run:   AdminRemoveFeed,
	}
)

func init() {
	adminCmd.AddCommand(adminRemoveFeedCmd)

	adminRemoveFeedCmd.Flags().StringVarP(&url, "url", "u", "", "base url of the site feed to be removed")
	adminRemoveFeedCmd.Flags().StringVarP(&feedUrl, "feedUrl", "f", "", "feed URL of the site to be removed")
}

func AdminRemoveFeed(cmd *cobra.Command, a []string) {
	fmt.Fprintf(os.Stderr, "removing feed...\n")
	if url == "" && feedUrl == "" {
		log.Fatalf("either the base url or the feed url should be specified")
	}

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

	var feeds *sql.Rows
	if feedUrl != "" {
		feeds, err = db.Query(fmt.Sprintf("SELECT id FROM %s WHERE feed = '%s' LIMIT 1", FEEDTABLE, feedUrl))
		if err != nil {
			log.Fatalf("error while looking up specified feed: [feedUrl] %s, %s", feedUrl, err)
		}
	} else if url != "" {
		feeds, err = db.Query(fmt.Sprintf("SELECT id FROM %s WHERE url = '%s' LIMIT 1", FEEDTABLE, url))
		if err != nil {
			log.Fatalf("error while looking up specified url: [url] %s, %s", url, err)
		}
	}

	rc := 0
	var id int64
	for feeds.Next() {
		rc += 1
		err = feeds.Scan(&id)
		if err != nil {
			log.Fatalf("error while parsing feed entry: %s", err)
		}
	}

	if rc == 0 {
		log.Fatalf("no rows were found matching the query")
	}

	delQ, err := db.Prepare(fmt.Sprintf("DELETE FROM %s WHERE id = %d", FEEDTABLE, id))
	if err != nil {
		log.Fatalf("error while prepping delete query: %s", err)
	}

	defer delQ.Close()

	_, err = delQ.Exec()
	if err != nil {
		log.Fatalf("error while deleting row from feeds table: %s", err)
	}

	log.Println("feed deleted")
}
