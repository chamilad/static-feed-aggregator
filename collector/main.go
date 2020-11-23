package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/mmcdole/gofeed"
)

// todo:
//  redo the package structure build a single binary with different profiles
//  each profile is implemented in different packages
//
func main() {
	c := flag.String("c", "config.yml", "configuration file")
	d := flag.String("d", "aggr.db", "items database")
	flag.Parse()

	fmt.Fprintf(os.Stderr, "starting feed collector...\n")

	// init db
	now := time.Now()
	db, err := sql.Open("sqlite3", *d)
	if err != nil {
		log.Fatalf("error while opening database: %s\n", err)
	}

	defer db.Close()

	itemsTable := fmt.Sprintf("items_%s", now.Format("200601"))
	stmt2, err := db.Prepare(fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s (id INTEGER PRIMARY KEY, timestamp INTEGER, title TEXT, body TEXT, url TEXT)", itemsTable))
	if err != nil {
		log.Fatalf("error while prepping create table: %s\n", err)
	}

	defer stmt2.Close()

	_, err = stmt2.Exec()
	if err != nil {
		log.Fatalf("error while executing create table: %s\n", err)
	}

	conf, err := common.LoadConfig(*c)
	if err != nil {
		log.Fatalf("error while loading configuration file: %s", err)
	}

	if len(conf.Aggr.Collector.Feeds) < 1 {
		log.Fatalf("no feeds are defined in the config")
	}

	// for each feed we have to check
	for i, feedConf := range conf.Aggr.Collector.Feeds {
		// fetch the feed items
		feed, err := feedFromUrl(feedConf.Feed)
		if err != nil {
			log.Printf("error while reading feeds from url %s: %s", feedConf.Feed, err)
			continue
		}

		fmt.Printf("title: %s\n", feed.Title)

		var lastRead int32
		if feedConf.LastRead == 0 {
			fmt.Fprintf(os.Stderr, "last_read not found for url %s, has to be inserted\n", feedConf.URL)
		} else {
			lastRead = feedConf.LastRead
		}

		// iterate the feed items untl timestamp is <= lastRead or until the end
		pCount := 0
		newLastRead := lastRead
		for _, item := range feed.Items {
			postTime := int32(0)
			if item.Updated != "" {
				postTime = int32((*item.UpdatedParsed).Unix())
			} else {
				postTime = int32((*item.PublishedParsed).Unix())
			}

			if postTime > lastRead {
				// item should be persisted
				addItem(db, itemsTable, postTime, item.Title, item.Description, item.Link)
				pCount++
				//fmt.Printf("\t%d: %s: %d -> %s => %s \n", i, item.GUID, postTime, item.Title, item.Description)

				// update lastRead if actual new feed items were found, with the latest timestamp
				if postTime > newLastRead {
					newLastRead = postTime
				}
			}
		}

		// if the newLastRead is updated, update the row in the db
		if pCount > 0 {
			log.Printf("%d new posts were processed from the feed", pCount)
			feedConf.LastRead = newLastRead
			conf.Aggr.Collector.Feeds[i] = feedConf
		} else {
			log.Printf("no new feeds were found for: %s\n", feedConf.URL)
		}
	}

	err = common.UpdateConfig(conf, *c)
	if err != nil {
		log.Fatalf("error while writing configuration file: %s", err)
	}
}

func feedFromUrl(u string) (feed *gofeed.Feed, err error) {
	fmt.Fprintf(os.Stderr, "feed: %s\n", u)
	resp, err := http.Get(u)
	if err != nil {
		log.Printf("error while talking to feed url: %s, %s", u, err)
		return nil, err
	}

	fp := gofeed.NewParser()
	feed, err = fp.Parse(resp.Body)
	if err != nil {
		log.Printf("error while getting the feed for feed %s, %s", feed, err)
		return nil, err
	}

	return feed, nil
}

func addItem(d *sql.DB, table string, timestamp int32, title, body, url string) error {
	stmt, err := d.Prepare(fmt.Sprintf("INSERT INTO %s(timestamp, title, body, url) VALUES(?,?,?,?)", table))
	if err != nil {
		log.Fatalf("error while prepping add feed: %s\n", err)
	}
	defer stmt.Close()

	_, err = stmt.Exec(timestamp, title, body, url)
	if err != nil {
		log.Fatalf("error while adding feed item: %s\n", err)
	}

	return nil
}
