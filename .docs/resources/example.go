package main

import (
	"fmt"
	"os"

	"github.com/johnmikee/yae"
)

type Config struct {
	Foo string `json:"foo"`
	Bar string `json:"bar"`
	Baz int    `json:"baz"`
}

func main() {
	var cfg Config
	err := yae.Get(
		yae.EnvType(yae.DEV),
		&yae.Env{
			Name:         "YAE",
			ConfigStruct: &cfg,
			Type:         yae.JSON,
		},
	)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	fmt.Println("Bar: ", cfg.Bar)
	fmt.Println("Baz: ", cfg.Baz)
	fmt.Println("Foo: ", cfg.Foo)
}
