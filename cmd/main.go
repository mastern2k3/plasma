package main

import (
	"encoding/json"
	"flag"
	"log"
	"os"

	"github.com/dop251/goja"
	"github.com/dop251/goja_nodejs/require"
)

const (
	bootstrapScript = `require(moduleName)`
)

var (
	file = flag.String("f", "", "js file to load")
)

func main() {

	flag.Parse()

	registry := new(require.Registry) // this can be shared by multiple runtimes

	runtime := goja.New()
	registry.Enable(runtime)

	f, err := os.Open(*file)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	var val goja.Value

	runtime.Set("moduleName", *file)

	if val, err = runtime.RunScript("bootstrapScript", bootstrapScript); err != nil {
		log.Fatal(err)
	}

	mapVal := val.Export().(map[string]interface{})

	jsonStr, err := json.Marshal(mapVal)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("val: `%s`", jsonStr)
}
