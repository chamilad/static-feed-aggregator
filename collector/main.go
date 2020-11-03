package main

import (
	"database/sql"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	_ "github.com/mattn/go-sqlite3"
	"github.com/mmcdole/gofeed"
	"gopkg.in/yaml.v2"
)

// 1. read feeds from file
// 2. generate content
// 3. store last read guid in the file
// 4. generate content for only the newer ones

type Config struct {
	Aggr struct {
		Collector struct {
			Feeds []struct {
				Name string `yaml:"name"`
				URL  string `yaml:"url"`
				Feed string `yaml:"feed"`
			} `yaml:"feeds"`
		} `yaml:"collector"`
	} `yaml:"aggregator"`
}

type FeedItem struct {
	Timestamp int32
	Title     string
	Body      string
}

func main() {
	c := flag.String("c", "config.yml", "configuration file")
	flag.Parse()

	fmt.Fprintf(os.Stderr, "starting feed collector...\n")

	// init db
	db, err := sql.Open("sqlite3", "./aggr.db")
	if err != nil {
		log.Fatalf("error while opening database: %s", err)
	}

	defer db.Close()

	stmt, _ := db.Prepare("CREATE TABLE IF NOT EXISTS feeds (id INTEGER PRIMARY KEY, url TEXT, last_read INTEGER)")
	defer stmt.Close()

	_, err = stmt.Exec()
	if err != nil {
		log.Fatalf("error while executing create table: %s", err)
	}

	// todo: check if file exists

	yc, err := ioutil.ReadFile(*c)
	if err != nil {
		log.Fatalf("error while reading config file: %s", err)
	}

	var conf Config
	err = yaml.Unmarshal(yc, &conf)
	if err != nil {
		log.Fatalf("error while parsing yaml config: %s", err)
	}

	if len(conf.Aggr.Collector.Feeds) < 1 {
		log.Fatalf("no feeds are defined in the config")
	}

	newFeeds := []*FeedItem{}

	// for each feed we have to check
	for _, feedConf := range conf.Aggr.Collector.Feeds {
		// fetch the feed items
		fp := gofeed.NewParser()
		feed, _ := fp.ParseURL(feedConf.Feed)
		fmt.Printf("\ntitle: %s\n", feed.Title)

		// get last read timestamp
		rows, err := db.Query(fmt.Sprintf("SELECT last_read FROM feeds WHERE url='%s'", feedConf.URL))
		if err != nil {
			log.Fatalf("error while getting the last read timestamp: %s", err)
		}
		defer rows.Close()

		var lastRead int32
		for rows.Next() {
			err = rows.Scan(&lastRead)
			if err != nil {
				log.Fatalf("error while parsing row")
			}
			fmt.Fprintf(os.Stderr, "lastRead: %d", lastRead)
		}

		// is there a lastRead timestamp in the db? if not, this is the first the time the feed is being read
		if lastRead == 0 {
			fmt.Fprintf(os.Stderr, "lastRead not found for url %s, has to be inserted\n", feedConf.URL)
		}

		// iterate the feed items untl timestamp is <= lastRead or until the end
		newLastRead := lastRead
		for i, item := range feed.Items {
			postTime := int32(0)
			if item.Updated != "" {
				postTime = int32((*item.UpdatedParsed).Unix())
			} else {
				postTime = int32((*item.PublishedParsed).Unix())
			}

			if postTime > lastRead {
				feedItem := &FeedItem{
					Timestamp: postTime,
					Title:     item.Title,
					Body:      item.Content,
				}

				newFeeds = append(newFeeds, feedItem)
				fmt.Printf("\t%d: %s: %d -> %s\n", i, item.GUID, postTime, item.Title)

				// update lastRead if actual new feed items were found, with the latest timestamp
				if postTime > newLastRead {
					newLastRead = postTime
				}
			}
		}

		// if the newLastRead is updated, update the row in the db
		if newLastRead != lastRead {
			// if lastRead was 0, then it should be an insert
			if lastRead == 0 {
				stmt, err := db.Prepare("INSERT INTO feeds(url, last_read) VALUES(?,?)")
				if err != nil {
					log.Fatalf("error while prepping insert stmt: %s\n", err)
				}

				_, err = stmt.Exec(feedConf.URL, newLastRead)
				if err != nil {
					log.Fatalf("error while executing insert stmt: %s\n", err)
				}
			} else { // else it should be an update
				stmt, err := db.Prepare("UPDATE feeds SET last_read = ? WHERE url = ?")
				if err != nil {
					log.Fatalf("error while prepping update stmt: %s\n", err)
				}

				_, err = stmt.Exec(newLastRead, feedConf.URL)
				if err != nil {
					log.Fatalf("error while executing update stmt: %s\n", err)
				}
			}
		} else {
			log.Printf("no new feeds were found for: %s\n", feedConf.URL)
		}
	}
}
