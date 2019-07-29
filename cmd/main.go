package main

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"

	"github.com/dop251/goja"
	"github.com/dop251/goja_nodejs/require"
)

const (
	precompileScript = `
		Babel = require("internal/babel6.min.js");
		newCode = Babel.transform(userCode, {
			sourceType: "script",
			presets: [
				['es2015', { "modules": false }],
				'stage-1'
			]
		});
		newCode.code
	`
	bootstrapScript = `
		require(moduleName)
	`
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

type LocalRegLoader struct {
	runtime *goja.Runtime
}

func (l LocalRegLoader) RegLoader(filename string) ([]byte, error) {

	var (
		val goja.Value
		err error
	)

	userCode, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	l.runtime.Set("userCode", string(userCode))

	if val, err = l.runtime.RunScript("precompileScript", precompileScript); err != nil {
		return nil, err
	}

	l.runtime.Set("userCode", nil)

	return []byte(val.Export().(string)), nil
}

func main() {

	flag.Parse()

	precompileRuntime := goja.New()
	precompileRegistry := new(require.Registry)
	precompileRegistry.Enable(precompileRuntime)
	loader := LocalRegLoader{precompileRuntime}

	runtime := goja.New()
	registry := require.NewRegistryWithLoader(loader.RegLoader)
	registry.Enable(runtime)

	objects := map[string]interface{}{}

	err := filepath.Walk(*directory, func(path string, f os.FileInfo, err error) error {

		if f.IsDir() {
			return nil
		}

		if strings.HasPrefix(path, "internal/") {
			return nil
		}

		ext := filepath.Ext(path)
		mod := strings.TrimSuffix(path, ext)

		switch ext {
		case ".js":

			obj, err := ResolveFileObject(path, runtime)
			if err != nil {
				log.Printf("error while resolving object in `%s`: %s", path, err)
				objects[mod] = err
			} else {
				objects[mod] = obj
			}

		case ".babelrc":
			return nil
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
