package main

import (
	"encoding/json"
	"flag"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"

	"github.com/dop251/goja"
	"github.com/dop251/goja_nodejs/require"
)

const (
	bootstrapScript = `require(moduleName)`
)

var (
	directory = flag.String("d", "", "directory to scan")
)

func ResolveFileObject(filename string, runtime *goja.Runtime) (map[string]interface{}, error) {

	var (
		val goja.Value
		err error
	)

	runtime.Set("moduleName", filename)

	if val, err = runtime.RunScript("bootstrapScript", bootstrapScript); err != nil {
		return nil, err
	}

	return val.Export().(map[string]interface{}), nil
}

func main() {

	flag.Parse()

	runtime := goja.New()

	registry := new(require.Registry)
	registry.Enable(runtime)

	objects := map[string]interface{}{}

	err := filepath.Walk(*directory, func(path string, f os.FileInfo, err error) error {

		if f.IsDir() {
			return nil
		}

		ext := filepath.Ext(path)
		mod := strings.TrimSuffix(path, ext)

		switch ext {
		case ".js":

			obj, err := ResolveFileObject(path, runtime)
			if err != nil {
				return err
			}

			objects[mod] = obj

		default:
			return errors.Errorf("unrecognized file type detected `%s`", path)
		}

		return nil
	})

	if err != nil {
		log.Fatal(err)
	}

	jsonStr, err := json.Marshal(objects)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("val: `%s`", jsonStr)
}
