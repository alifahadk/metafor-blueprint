package config

import (
	"encoding/json"
	"os"
)

type SystemConfig struct {
	Servers []Server `json:"servers"`
}

type Server struct {
	Name       string         `json:"name"`
	ThreadPool uint           `json:"threadpool"`
	QueueSize  uint           `json:"qsize"`
	APIs       map[string]API `json:"apis"`
}

type API struct {
	ProcessingRate     float64         `json:"processing_rate"`
	DownstreamServices []DownstreamAPI `json:"downstream_services"`
}

type DownstreamAPI struct {
	Source  string `json:"source"`
	Target  string `json:"target"`
	API     string `json:"api"`
	Timeout int    `json:"timeout"`
	Retry   int    `json:"retry"`
}

func LoadConfig(path string) (SystemConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return SystemConfig{}, err
	}
	var cfg SystemConfig
	err = json.Unmarshal(data, &cfg)
	return cfg, err
}
