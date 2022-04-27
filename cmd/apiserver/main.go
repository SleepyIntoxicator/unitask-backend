package main

import (
	"backend/internal/apiserver"
	"flag"
)

var (
	configPath string
)

func init() {
	flag.StringVar(&configPath, "config-path", "configs/", "path to config file")

}

func main() {
	flag.Parse()
	apiserver.Run(configPath)
}
