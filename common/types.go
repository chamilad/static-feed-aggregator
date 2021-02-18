package common

type Config struct {
	Database string `yaml:"database"`
	Aggr     struct {
		Collector struct {
			Feeds []struct {
				Title    string `yaml:"title"`
				URL      string `yaml:"url"`
				Feed     string `yaml:"feed"`
				LastRead int32  `yaml:"last_read"`
			} `yaml:"feeds"`
		} `yaml:"collector"`
		Renderer struct {
			Collection struct {
				Days int `yaml:"days"`
			} `yaml:"collection"`
			Site struct {
				Location string `yaml:"location"`
				FileName string `yaml:"file_name"`
			} `yaml:"site"`
		} `yaml:"renderer"`
	} `yaml:"aggregator"`
}

type Post struct {
	Title     string `yaml:"title"`
	URL       string `yaml:"url"`
	Published string `yaml:"published"`
	Fragment  string `yaml:"frag"`
}

type Posts struct {
	Posts []*Post `yaml:"posts"`
}
