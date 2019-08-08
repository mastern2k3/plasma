package web

import (
	"fmt"
	"io"
	"net/http"
	"text/template"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

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

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	e.Renderer = &Template{
		templates: HomeTemplate,
	}

	e.GET("/", func(c echo.Context) error {

		mod := HomeModel{}

		for _, o := range dir {
			mod.Objects = append(mod.Objects, HomeObjectModel{
				Path:  o.Path,
				Error: o.Error,
			})
		}

		return c.Render(http.StatusOK, "home", mod)
	})

	e.GET("/*", func(c echo.Context) error {

		path := c.Param("*")

		cached, has := dir[path]

		if !has {
			return c.String(http.StatusNotFound, fmt.Sprintf("could not find `%s`", path))
		}

		if cached.Error != nil {
			return c.String(http.StatusFailedDependency, fmt.Sprintf("error resolving `%s`", path))
		}

		if _, has := c.QueryParams()["meta"]; has {
			return c.JSONPretty(http.StatusOK, cached, "  ")
		}

		return c.String(http.StatusOK, cached.Cached)
	})

	e.Logger.Fatal(e.Start(":1323"))
}
