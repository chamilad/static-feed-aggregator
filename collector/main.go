package main

import (
	"fmt"
	"os"

	"github.com/mmcdole/gofeed"
)

func main() {
	fmt.Fprintf(os.Stderr, "starting feed collector...\n")

	feeds := []string{"https://monkeyseemonkeykill.wordpress.com/feed", "http://feeds.twit.tv/twit.xml"}
	for _, feedVal := range feeds {
		fp := gofeed.NewParser()
		feed, _ := fp.ParseURL(feedVal)
		fmt.Printf("title: %s\n", feed.Title)
		for i, item := range feed.Items {
			// guid could be used as an identifier for the last read item
			fmt.Printf("\t%d: %s -> %s\n", i, item.GUID, item.Title)
		}
	}
}
