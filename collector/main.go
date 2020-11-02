package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"

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
				Name     string `yaml:"name"`
				URL      string `yaml:"url"`
				Feed     string `yaml:"feed"`
				LastRead string `yaml:"lastRead"`
			} `yaml:"feeds"`
		} `yaml:"collector"`
	} `yaml:"aggregator"`
}

func main() {
	c := flag.String("c", "config.yml", "configuration file")
	flag.Parse()

	fmt.Fprintf(os.Stderr, "starting feed collector...\n")

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

	for _, feedConf := range conf.Aggr.Collector.Feeds {
		fp := gofeed.NewParser()
		feed, _ := fp.ParseURL(feedConf.Feed)
		fmt.Printf("title: %s\n", feed.Title)
		for i, item := range feed.Items {
			// guid could be used as an identifier for the last read item
			fmt.Printf("\t%d: %s -> %s\n", i, item.GUID, item.Title)
		}
	}
}
