package main

import (
	core "github.com/weblfe/protoc-gen-api/pkg/app"
	"github.com/weblfe/protoc-gen-api/pkg/generators"
)

var (
	app = core.NewProtocPluginApp()
	name = `protoc-gen-api`
)

func main() {
	if err := app.Run(); err != nil {
		//panic(err)
	}
}

func init() {
	app.Register(generators.NewApiGenerator())
}

func GetName() string {
		return name
}
