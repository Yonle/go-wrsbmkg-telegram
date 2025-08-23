package main

import (
	"fmt"
	"os"

	"github.com/goccy/go-yaml"
)

type RegionFilter struct {
	Coords struct {
		Latitude  float64 `yaml:"Latitude"`
		Longitude float64 `yaml:"Longitude"`
	} `yaml:"Coordinates"`
	MaxDistance  float64 `yaml:"MaxDistance"`
	MinMagnitude float64 `yaml:"MinMagnitude"`
	Ignore       bool    `yaml:"Ignore"`
}

var config struct {
	Token         string         `yaml:"BOT_TOKEN"`
	ChatID        int            `yaml:"CHAT_ID"`
	MinMag        float64        `yaml:"MIN_MAG"`
	RegionsFilter []RegionFilter `yaml:"REGIONS_FILTER"`
	MsgMemoryDir  string         `yaml:"MSG_MEMORY_DIR"`
}

func ReadConfig(filename string) {
	data, err := os.ReadFile(filename)
	if err != nil {
		if os.IsNotExist(err) {
			panic("Config file does not exist.")
		}

		panic(fmt.Sprintf("error when reading %s: %s", filename, err))
	}

	if err := yaml.Unmarshal(data, &config); err != nil {
		panic(fmt.Sprintf("error when parsing %s: %s", filename, err))
	}
}
