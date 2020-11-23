package cmd

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"time"

	"github.com/chamilad/sinhala-blog-aggregator/common"
	_ "github.com/mattn/go-sqlite3"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

var (
	renderCmd = &cobra.Command{
		Use:   "render",
		Short: "renders the output",
		Long:  "",
		Run:   Render,
	}
)

type Post struct {
	Title     string `yaml:"title"`
	URL       string `yaml:"url"`
	Published string `yaml:"published"`
	Fragment  string `yaml:"frag"`
}

type Posts struct {
	Posts []*Post `yaml:"posts"`
}

func init() {
	rootCmd.AddCommand(renderCmd)
}

func Render(cmd *cobra.Command, a []string) {
	fmt.Fprintf(os.Stderr, "starting feed renderer...\n")
	db, err := common.OpenDb(d)
	if err != nil {
		log.Fatalf("error while opening database: %s\n", err)
	}

	defer db.Close()

	conf, err := common.LoadConfig(c)
	if err != nil {
		log.Fatalf("error while loading configuration: %s", err)
	}

	now := time.Now()
	itemsTable := fmt.Sprintf("items_%s", now.Format("200601"))
	timeConstraint := now.Unix() - (60 * 60 * 24 * int64(conf.Aggr.Renderer.Collection.Days)) // x days before today

	// SELECT * FROM {itemsTable} WHERE timestamp >= {today-10d.12AM} ORDER BY timestamp DSC
	posts, err := db.Query(fmt.Sprintf("SELECT id, timestamp, title, body, url FROM %s WHERE timestamp >= %d ORDER BY timestamp DESC", itemsTable, timeConstraint))
	if err != nil {
		log.Fatalf("error while retrieving posts: %s", err)
	}

	var rendPosts []*Post

	for posts.Next() {
		var id int
		var timestamp int64
		var title string
		var body string
		var url string

		err = posts.Scan(&id, &timestamp, &title, &body, &url)
		if err != nil {
			log.Fatalf("error while parsing post entry: %s", err)
		}

		p := &Post{
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

	yp := &Posts{
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

	//fmt.Fprintf(os.Stderr, "%s", string(d))

	//t, err := template.ParseFiles("posts.md.tpl")
	//if err != nil {
	//log.Fatalf("error while parsing template file: %s", err)
	//}

	//tData := struct {
	//Posts []*Post
	//}{
	//Posts: rendPosts,
	//}

	//err = t.Execute(os.Stdout, tData)
	//if err != nil {
	//log.Fatalf("error while rendering template: %s", err)
	//}

	fmt.Fprintf(os.Stderr, "all done")
}
