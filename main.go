package main

import (
	"github.com/joho/godotenv"
	core "github.com/weblfe/protoc-gen-api/pkg/app"
	"github.com/weblfe/protoc-gen-api/pkg/generators"
		"os"
)

var (
	version = `v1.0.0`
	name = `protoc-gen-api`
	app  = core.NewProtocPluginApp()
)

func main() {

	if err := app.Run(); err != nil {
		panic(err)
	}
}

func init() {
	if err := godotenv.Load(); err != nil && !os.IsNotExist(err) {
		panic(err)
	}
	core.SetName(name)
	core.SetVersion(version)
	app.Register(generators.NewApiGenerator())
}

