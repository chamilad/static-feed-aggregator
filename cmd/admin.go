package cmd

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/chamilad/sinhala-blog-aggregator/common"
	"github.com/olekukonko/tablewriter"
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

	adminListFeedsCmd = &cobra.Command{
		Use:   "listFeeds",
		Short: "list the feeds that are being aggregated by this instance",
		Long:  "",
		Run:   AdminListFeeds,
	}

	title   string
	url     string
	feedUrl string
	email   string

	adminAddFeedCmd = &cobra.Command{
		Use:   "addFeed",
		Short: "add a new feed to be aggregated",
		Long:  "",
		Run:   AdminAddFeed,
	}

	adminRemoveFeedCmd = &cobra.Command{
		Use:   "removeFeed",
		Short: "remove a feed from being aggregated",
		Long:  "",
		Run:   AdminRemoveFeed,
	}
)

func init() {
	rootCmd.AddCommand(adminCmd)
	adminCmd.AddCommand(adminResetDatesCmd)
	adminCmd.AddCommand(adminListFeedsCmd)
	adminCmd.AddCommand(adminAddFeedCmd)
	adminCmd.AddCommand(adminRemoveFeedCmd)

	adminAddFeedCmd.Flags().StringVarP(&title, "title", "t", "", "title of the site")
	adminAddFeedCmd.Flags().StringVarP(&url, "url", "u", "", "base url of the site")
	adminAddFeedCmd.Flags().StringVarP(&feedUrl, "feedUrl", "f", "", "feed URL of the site")
	adminAddFeedCmd.Flags().StringVarP(&email, "email", "e", "", "email address of the requester")
	adminAddFeedCmd.MarkFlagRequired("title")
	adminAddFeedCmd.MarkFlagRequired("url")
	adminAddFeedCmd.MarkFlagRequired("feedUrl")
	adminAddFeedCmd.MarkFlagRequired("email")
}

func Admin(cmd *cobra.Command, a []string) {
	fmt.Fprintf(os.Stderr, "a tool to manage the config.yaml")
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

	for feeds.Next() {
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

	w.Render()
	fmt.Fprintf(os.Stderr, "all done")

}

func AdminAddFeed(cmd *cobra.Command, a []string) {
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

func AdminRemoveFeed(cmd *cobra.Command, a []string) {
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
