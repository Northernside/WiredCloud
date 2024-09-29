package main

import (
	"wiredcloud/api"
	"wiredcloud/modules/env"
)

func main() {
	env.LoadEnvFile()
	api.StartWebServer()
}
