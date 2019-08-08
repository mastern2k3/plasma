package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"flag"
	"hash/fnv"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	_ "github.com/pkg/errors"

	"github.com/dop251/goja"
	"github.com/dop251/goja_nodejs/require"
	"github.com/gobuffalo/packr/v2"

	"github.com/mastern2k3/plasma"
	"github.com/mastern2k3/plasma/model"
	u "github.com/mastern2k3/plasma/util"
	"github.com/mastern2k3/plasma/web"
)

const (
	precompileScript = `
		Babel = require("babel6.min.js");
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

type BoxRegLoader struct {
	box *packr.Box
}

func (l BoxRegLoader) RegLoader(filename string) ([]byte, error) {
	return l.box.Find(filename)
}

func main() {

	flag.Parse()

	internalsBox := packr.New("Internals", "./internal")

	precompileRuntime := goja.New()
	internalsBoxLoader := BoxRegLoader{internalsBox}
	precompileRegistry := require.NewRegistryWithLoader(internalsBoxLoader.RegLoader)
	precompileRegistry.Enable(precompileRuntime)
	loader := LocalRegLoader{precompileRuntime}

	runtime := goja.New()
	registry := require.NewRegistryWithLoader(loader.RegLoader)
	registry.Enable(runtime)

	objects := map[string]model.DataObject{}

	err := filepath.Walk(*directory, func(path string, f os.FileInfo, err error) error {

		if f.IsDir() {
			return nil
		}

		return DigestFile(path, objects, runtime)
	})

	if err != nil {
		u.Logger.Fatal(err)
	}

	changedFiles := make(chan string)

	go func() {
		if err := plasma.StartWatching(context.Background(), *directory, changedFiles); err != nil {
			u.Logger.WithError(err).Fatal("error while watching")
		}
	}()

	go func() {
		for path := range changedFiles {
			if err := DigestFile(filepath.Clean(path), objects, runtime); err != nil {
				u.Logger.WithError(err).Fatal("error while digesting file changes")
			}
		}
	}()

	web.StartServer(objects)
}

func DigestFile(path string, objects model.ObjectDirectory, runtime *goja.Runtime) error {

	ext := filepath.Ext(path)
	mod := strings.TrimSuffix(path, ext)

	var dat interface{}

	switch ext {
	case ".js":

		obj, err := ResolveFileObject(path, runtime)
		if err != nil {

			u.Logger.WithError(err).Errorf("error while resolving object in `%s`", path)

			objects[mod] = model.DataObject{
				Path:  mod,
				Error: err,
			}

			return nil
		}

		dat = obj

	default:
		u.Logger.Warningf("unrecognized file type detected `%s`", path)
		return nil
	}

	jsonBytes, err := json.Marshal(dat)
	if err != nil {
		return err
	}

	hash := fnv.New128a()
	hash.Write(jsonBytes)
	hashBytes := hash.Sum(make([]byte, hash.Size()))
	hashString := base64.StdEncoding.EncodeToString(hashBytes)

	objects[mod] = model.DataObject{
		Path:   mod,
		Data:   dat,
		Hash:   hashString,
		Cached: string(jsonBytes),
	}

	return nil
}
