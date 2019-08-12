package javascript

import (
	"strings"

	"github.com/dop251/goja"
	"github.com/dop251/goja_nodejs/require"
	"github.com/gobuffalo/packr/v2"
)

const (
	precompileScript = `
		Babel = require("babel6.min.js");
		newCode = Babel.transform(userCode, {
			sourceType: "module",
			presets: [
				['es2015'],
				'stage-1'
			],
		});
		newCode.code
	`
)

var (
	internalsBox = packr.New("Internals", "./internal")
)

type Precompiler struct {
	runtime *goja.Runtime
}

func NewPrecompiler() *Precompiler {

	runtime := goja.New()

	internalsRegistry := require.NewRegistryWithLoader(func(filename string) ([]byte, error) {
		return internalsBox.Find(filename)
	})

	internalsRegistry.Enable(runtime)

	return &Precompiler{
		runtime: runtime,
	}
}

func (p *Precompiler) Precompile(input []byte) ([]byte, error) {

	var (
		val goja.Value
		err error
	)

	p.runtime.Set("userCode", string(input))

	if val, err = p.runtime.RunScript("precompileScript", precompileScript); err != nil {
		return nil, err
	}

	p.runtime.Set("userCode", nil)

	newJs := val.Export().(string)

	newJs = strings.Replace(newJs, "\"use strict\";", "", 1)

	return []byte(newJs), nil
}
