package common

import (
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

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
