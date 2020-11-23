package common

import (
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

type Config struct {
	Aggr struct {
		Collector struct {
			Feeds []struct {
				Title    string `yaml:"title"`
				URL      string `yaml:"url"`
				Feed     string `yaml:"feed"`
				LastRead int32  `yaml:"last_read"`
			} `yaml:"feeds"`
		} `yaml:"collector"`
	} `yaml:"aggregator"`
}

func LoadConfig(f string) (c *Config, err error) {
	// todo: check if file exists

	yc, err := ioutil.ReadFile(f)
	if err != nil {
		return nil, err
	}

	var conf Config
	err = yaml.Unmarshal(yc, &conf)
	if err != nil {
		return nil, err
	}

	return &conf, nil
}

func UpdateConfig(c *Config, f string) error {
	// write the state back to yaml
	cf, err := yaml.Marshal(c)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(f, []byte(cf), 0644)
	if err != nil {
		return err
	}

	return nil
}
