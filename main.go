package main

import (
	"log"

	"github.com/ayumi-otosaka-314/brec-pp/config"
	"github.com/ayumi-otosaka-314/brec-pp/registry"
)

func main() {
	conf, err := config.New()
	if err != nil {
		log.Fatalln(err)
	}

	r := registry.New(conf)
	r.NewServer().Serve()
	r.CleanUp()
}
