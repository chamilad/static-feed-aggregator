package main

import (
	"database/sql"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

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
				Title string `yaml:"title"`
				URL   string `yaml:"url"`
				Feed  string `yaml:"feed"`
			} `yaml:"feeds"`
		} `yaml:"collector"`
	} `yaml:"aggregator"`
}

func main() {
	c := flag.String("c", "config.yml", "configuration file")
	flag.Parse()

	fmt.Fprintf(os.Stderr, "starting feed collector...\n")

	// init db
	now := time.Now()
	db, err := sql.Open("sqlite3", "./aggr.db")
	if err != nil {
		log.Fatalf("error while opening database: %s\n", err)
	}

	defer db.Close()

	stmt1, err := db.Prepare("CREATE TABLE IF NOT EXISTS feeds (url TEXT PRIMARY KEY, title TEXT, feed TEXT, last_read INTEGER)")
	if err != nil {
		log.Fatalf("error while prepping create table: %s\n", err)
	}

	defer stmt1.Close()

	_, err = stmt1.Exec()
	if err != nil {
		log.Fatalf("error while executing create table: %s\n", err)
	}

	itemsTable := fmt.Sprintf("items_%s", now.Format("200601"))
	stmt2, err := db.Prepare(fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s (id INTEGER PRIMARY KEY, timestamp INTEGER, title TEXT, body TEXT)", itemsTable))
	if err != nil {
		log.Fatalf("error while prepping create table: %s\n", err)
	}

	defer stmt2.Close()

	_, err = stmt2.Exec()
	if err != nil {
		log.Fatalf("error while executing create table: %s\n", err)
	}

	// todo: check if file exists

	yc, err := ioutil.ReadFile(*c)
	if err != nil {
		log.Fatalf("error while reading config file: %s\n", err)
	}

	// todo: feeds should be read from the feeds table
	//
	var conf Config
	err = yaml.Unmarshal(yc, &conf)
	if err != nil {
		log.Fatalf("error while parsing yaml config: %s\n", err)
	}

	if len(conf.Aggr.Collector.Feeds) < 1 {
		log.Fatalf("no feeds are defined in the config")
	}

	//	newFeeds := []*FeedItem{}

	// for each feed we have to check
	for _, feedConf := range conf.Aggr.Collector.Feeds {
		// fetch the feed items
		feed, err := feedFromUrl(feedConf.Feed)

		//feed, _ := fp.ParseURL(feedConf.Feed)
		fmt.Printf("\ntitle: %s\n", feed.Title)

		// get last read timestamp
		rows, err := db.Query(fmt.Sprintf("SELECT last_read FROM feeds WHERE url='%s'", feedConf.URL))
		if err != nil {
			log.Fatalf("error while getting the last read timestamp: %s\n", err)
		}
		defer rows.Close()

		var lastRead int32
		for rows.Next() {
			err = rows.Scan(&lastRead)
			if err != nil {
				log.Fatalf("error while parsing row")
			}
			fmt.Fprintf(os.Stderr, "lastRead: %d\n", lastRead)
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
				// item should be persisted
				addItem(db, itemsTable, postTime, item.Title, item.Description)
				fmt.Printf("\t%d: %s: %d -> %s => %s \n", i, item.GUID, postTime, item.Title, item.Description)

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

				defer stmt.Close()

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

func addItem(d *sql.DB, table string, timestamp int32, title, body string) error {
	stmt, err := d.Prepare(fmt.Sprintf("INSERT INTO %s(timestamp, title, body) VALUES(?,?,?)", table))
	if err != nil {
		log.Fatalf("error while prepping add feed: %s\n", err)
	}
	defer stmt.Close()

	_, err = stmt.Exec(timestamp, title, body)
	if err != nil {
		log.Fatalf("error while adding feed item: %s\n", err)
	}

	return nil
}
