package cmd

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/chamilad/sinhala-blog-aggregator/common"
	"github.com/spf13/cobra"
)

var (
	adminAddFeedCmd = &cobra.Command{
		Use:   "addFeed",
		Short: "add a new feed to be aggregated",
		Long:  "",
		Run:   AdminAddFeed,
	}
)

func init() {
	adminCmd.AddCommand(adminAddFeedCmd)

	adminAddFeedCmd.Flags().StringVarP(&title, "title", "t", "", "title of the site")
	adminAddFeedCmd.Flags().StringVarP(&url, "url", "u", "", "base url of the site")
	adminAddFeedCmd.Flags().StringVarP(&feedUrl, "feedUrl", "f", "", "feed URL of the site")
	adminAddFeedCmd.Flags().StringVarP(&email, "email", "e", "", "email address of the requester")
	adminAddFeedCmd.MarkFlagRequired("title")
	adminAddFeedCmd.MarkFlagRequired("url")
	adminAddFeedCmd.MarkFlagRequired("feedUrl")
	adminAddFeedCmd.MarkFlagRequired("email")

}

func AdminAddFeed(cmd *cobra.Command, a []string) {
	fmt.Fprintf(os.Stderr, "adding feed...\n")

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

	stmt, err := db.Prepare(fmt.Sprintf("INSERT INTO %s(title, url, feed, email, added, last_read) VALUES(?,?,?,?,?,?)", FEEDTABLE))
	if err != nil {
		log.Fatalf("error while prepping add feed: %s\n", err)
	}
	defer stmt.Close()

	addedTime := time.Now().Unix()
	_, err = stmt.Exec(title, url, feedUrl, email, addedTime, 0)
	if err != nil {
		log.Fatalf("error while adding feed: %s\n", err)
	}

	log.Printf("post added with title: %s", title)
}
