package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

func main() {
	db, err := sql.Open("sqlite3", "../collector/aggr.db")
	if err != nil {
		log.Fatalf("error while opening database: %s\n", err)
	}

	defer db.Close()

	now := time.Now()
	itemsTable := fmt.Sprintf("items_%s", now.Format("200601"))
	timeConstraint := now.Unix() - (60 * 60 * 24 * 10) // 10 days before today

	// SELECT * FROM {itemsTable} WHERE timestamp >= {today-10d.12AM}
	posts, err := db.Query(fmt.Sprintf("SELECT * FROM %s WHERE timestamp >= %d", itemsTable, timeConstraint))
	if err != nil {
		log.Fatalf("error while retrieving posts: %s", err)
	}

	for posts.Next() {
		var id int
		var timestamp int64
		var title string
		var body string

		err = posts.Scan(&id, &timestamp, &title, &body)
		if err != nil {
			log.Fatalf("error while parsing post entry: %s", err)
		}

		fmt.Fprintf(
			os.Stderr,
			"==== post: %s, %s published on %s\n\n",
			title,
			//body[0:int(math.Max(10, float64(len(body))))],
			body,
			time.Unix(timestamp, 0).Format("2006-01-01"),
		)
	}
}
