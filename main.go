package main

import (
	"github.com/ayumi-otosaka-314/brec-pp/config"
	"github.com/ayumi-otosaka-314/brec-pp/registry"
)

func main() {
	r := registry.New(&config.Root{
		// Insert your config here
	})
	r.NewServer().Serve()
	r.CleanUp()
}
