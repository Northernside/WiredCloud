package main

import (
	"runtime"
	"time"

	"wiredcloud/api"
	"wiredcloud/modules/env"
	"wiredcloud/modules/sqlite"
)

func main() {
	env.LoadEnvFile()

	go func() {
		for {
			runtime.GC()
			time.Sleep(5 * time.Second)
		}
	}()

	sqlite.Init()
	api.StartWebServer()
}
