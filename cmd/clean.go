package cmd

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"strconv"
	"time"

	"github.com/chamilad/sinhala-blog-aggregator/common"
	_ "github.com/mattn/go-sqlite3"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

var (
	cleanCmd = &cobra.Command{
		Use:   "cleanup",
		Short: "cleanup older items from the database",
		Long:  "",
		Run:   Cleanup,
	}
)

func init() {
	rootCmd.AddCommand(cleanCmd)
}

func Cleanup(cmd *cobra.Command, a []string) {
	fmt.Fprintf(os.Stderr, "%s not implemented yet\n", cmd.Name())
	conf, err := common.LoadConfig(c)
	if err != nil {
		log.Fatalf("error while loading configuration: %s", err)
	}

	// to be run once a month on the month start
	// 1. get all the posts from last month
	// 2. write an archive yaml file to data/archives/month-year.md
	// 3. ???
	// 4. profit

	// get last month end timestamp
	t := time.Now()
	y, m, _ := t.Date()
	fmt.Printf("%d, %d\n", y, m)
	sotm := time.Date(y, m, 1, 0, 0, 0, 0, t.Location())
	fmt.Printf("%s\n", strconv.FormatInt(sotm.Unix(), 10))
	eolmtf := time.Unix(sotm.Unix()-(1), 0)
	fmt.Printf("%d, %s\n", eolmtf.Unix(), eolmtf.Format(time.RFC1123Z))

	solm := time.Date(y, m-1, 1, 0, 0, 0, 0, t.Location())
	solmtf := time.Unix(solm.Unix(), 0)
	fmt.Printf("%d, %s\n", solmtf.Unix(), solmtf.Format(time.RFC1123Z))

	db, err := common.OpenDb(conf.Database)
	if err != nil {
		log.Fatalf("error while opening database: %s\n", err)
	}

	defer db.Close()

	//itemsTable := fmt.Sprintf("items_%s", solm.Format("200601"))
	itemsTable := "items"

	sq := fmt.Sprintf("SELECT id, timestamp, title, body, url FROM %s WHERE timestamp <= %d AND timestamp >= %d ORDER BY timestamp DESC", itemsTable, eolmtf.Unix(), solmtf.Unix())
	fmt.Fprintf(os.Stderr, "query: %s\n", sq)
	pflm, err := db.Query(sq)
	if err != nil {
		// ex: no such table
		log.Fatalf("error while retrieving last month's posts: %s\n", err)
	}

	var rendPosts []*common.Post

	for pflm.Next() {
		var id int
		var timestamp int64
		var title string
		var body string
		var url string

		err = pflm.Scan(&id, &timestamp, &title, &body, &url)
		if err != nil {
			log.Fatalf("error while parsing post entry: %s", err)
		}

		p := &common.Post{
			Title:     title,
			URL:       url,
			Published: fmt.Sprintf("%d", timestamp),
			Fragment:  body,
		}

		rendPosts = append(rendPosts, p)

		fmt.Fprintf(
			os.Stderr,
			"==== post: %s, %s published on %s\n\n",
			title,
			//body[0:int(math.Max(10, float64(len(body))))],
			body,
			time.Unix(timestamp, 0).Format("2006-01-01"),
		)
	}

	if len(rendPosts) < 1 {
		log.Printf("no new posts were found to render, aborting...")
		os.Exit(0)
	}

	yp := &common.Posts{
		Posts: rendPosts,
	}

	d, err := yaml.Marshal(yp)
	if err != nil {
		log.Fatalf("error while rendering yaml output: %s", err)
	}

	dtFile := path.Join(conf.Aggr.Renderer.Site.Location, "data", conf.Aggr.Renderer.Site.FileName)
	err = ioutil.WriteFile(dtFile, d, 0644)
	if err != nil {
		log.Fatalf("error while writing into %s: %s", dtFile, err)
	}

}
