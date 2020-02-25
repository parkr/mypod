package mypod

import (
	"encoding/json"
	"os"
)

type Config struct {
	BaseURL     string
	Title       string
	Description string
	Link        string
	Language    string
	Copyright   string
	Author      string
	Subtitle    string
	Summary     string
	Owner       struct {
		Name  string
		Email string
	}
	Image string
}

func ReadConfig(filepath string) (Config, error) {
	f, err := os.Open(filepath)
	if err != nil {
		return Config{}, err
	}
	defer f.Close()

	conf := Config{}
	err = json.NewDecoder(f).Decode(&conf)
	return conf, err
}
