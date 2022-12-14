package main

import (
	"fmt"

	"gopkg.in/yaml.v3"

	"github.com/asmyasnikov/ydb-docker/internal/config"
)

func main() {
	cfg, err := config.New(false)
	if err != nil {
		panic(err)
	}
	bb, err := yaml.Marshal(cfg)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(bb))
}
