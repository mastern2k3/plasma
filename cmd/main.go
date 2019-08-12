package main

import (
	"bytes"
	"context"
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

	"github.com/mastern2k3/plasma"
	"github.com/mastern2k3/plasma/javascript"
	"github.com/mastern2k3/plasma/model"
	u "github.com/mastern2k3/plasma/util"
	"github.com/mastern2k3/plasma/web"
)

const (
	bootstrapScript = `
		require(moduleName)
	`
)

var (
	directory       = flag.String("d", "", "directory to scan")
	debugPrecompile = flag.String("dp", "", "debug precompile")
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

type PrecompilingLoader struct {
	precompiler *javascript.Precompiler
}

func (l PrecompilingLoader) RegLoader(filename string) ([]byte, error) {

	if filepath.Ext(filename) != ".js" {
		filename = filename + ".js"
	}

	userCode, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	return l.precompiler.Precompile(userCode)
}

func main() {

	flag.Parse()

	if *debugPrecompile != "" {

		precompiler := javascript.NewPrecompiler()

		userCode, err := ioutil.ReadFile(*debugPrecompile)
		if err != nil {
			panic(err)
		}

		newJs, err := precompiler.Precompile(userCode)
		if err != nil {
			panic(err)
		}

		u.Logger.Info(string(newJs))

		return
	}

	runtime := goja.New()

	precompiler := javascript.NewPrecompiler()

	loader := PrecompilingLoader{precompiler}
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

			runtime := goja.New()

			loader := PrecompilingLoader{precompiler}
			registry := require.NewRegistryWithLoader(loader.RegLoader)
			registry.Enable(runtime)

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
				Path:         mod,
				Error:        err,
				ErrorMessage: err.Error(),
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

	if oldObj, has := objects[mod]; has {

		if bytes.Equal(hashBytes, oldObj.Hash) {
			u.Logger.Infof("digested object equal to old record, ignoring")
			return nil
		} else {
			u.Logger.Infof("digested object different from old record, updating")
		}
	}

	objects[mod] = model.DataObject{
		Path:   mod,
		Data:   dat,
		Hash:   hashBytes,
		Cached: jsonBytes,
	}

	return nil
}
