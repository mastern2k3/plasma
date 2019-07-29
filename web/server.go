package web

import (
	"fmt"
	"io"
	"net/http"
	"text/template"

	"github.com/labstack/echo/v4"

	"github.com/mastern2k3/plasma/model"
)

type Template struct {
	templates *template.Template
}

func (t *Template) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
}

func StartServer(dir model.ObjectDirectory) {

	e := echo.New()

	e.Renderer = &Template{
		templates: HomeTemplate,
	}

	e.GET("/", func(c echo.Context) error {

		mod := HomeModel{}

		for _, o := range dir {
			mod.Objects = append(mod.Objects, HomeObjectModel{
				Path: o.Path,
			})
		}

		return c.Render(http.StatusOK, "home", mod)
	})

	e.GET("/*", func(c echo.Context) error {

		path := c.Param("*")

		cached, has := dir[path]

		if !has {
			return c.String(http.StatusNotFound, fmt.Sprintf("Could not find `%s`", path))
		}

		return c.String(http.StatusOK, cached.Cached)
	})

	e.Logger.Fatal(e.Start(":1323"))
}
