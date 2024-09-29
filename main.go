package main

import (
	"runtime"
	"time"
	"wiredcloud/api"
	"wiredcloud/modules/env"
)

func main() {
	env.LoadEnvFile()

	go func() {
		for {
			runtime.GC()
			time.Sleep(5 * time.Second)
		}
	}()

	api.StartWebServer()
}
