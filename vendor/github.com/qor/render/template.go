package render

import (
	"bytes"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"path/filepath"

	"github.com/cryptix/go/logging"
	"github.com/pkg/errors"
)

// Template template struct
type Template struct {
	render  *Render
	layout  string
	funcMap template.FuncMap
}

// FuncMap get func maps from tmpl
func (tmpl *Template) funcMapMaker(req *http.Request, writer http.ResponseWriter) template.FuncMap {
	var funcMap = template.FuncMap{}

	for key, fc := range tmpl.render.funcMaps {
		funcMap[key] = fc
	}

	if tmpl.render.Config.FuncMapMaker != nil {
		for key, fc := range tmpl.render.Config.FuncMapMaker(tmpl.render, req, writer) {
			funcMap[key] = fc
		}
	}

	for key, fc := range tmpl.funcMap {
		funcMap[key] = fc
	}
	return funcMap
}

// Funcs register Funcs for tmpl
func (tmpl *Template) Funcs(funcMap template.FuncMap) *Template {
	tmpl.funcMap = funcMap
	return tmpl
}

// Render render tmpl
func (tmpl *Template) Render(templateName string, obj interface{}, request *http.Request, writer http.ResponseWriter) (template.HTML, error) {
	var (
		content []byte
		t       *template.Template
		err     error
		funcMap = tmpl.funcMapMaker(request, writer)
		render  = func(name string, objs ...interface{}) (template.HTML, error) {
			var (
				err           error
				renderObj     interface{}
				renderContent []byte
			)

			if len(objs) == 0 {
				// default obj
				renderObj = obj
			} else {
				// overwrite obj
				for _, o := range objs {
					renderObj = o
					break
				}
			}

			renderContent, err = tmpl.findTemplate(name)
			if err != nil {
				return "yieldFindErr", errors.Wrapf(err, "failed to find template: %v", templateName)
			}
			var partialTemplate *template.Template
			var result bytes.Buffer
			partialTemplate, err = template.New(filepath.Base(name)).Funcs(funcMap).Parse(string(renderContent))
			if err != nil {
				return "yieldParseERR", err
			}
			err = partialTemplate.Execute(&result, renderObj)
			if err != nil {
				return "yieldExecERROR", err
			}
			return template.HTML(result.String()), nil
		}
	)

	// funcMaps
	funcMap["render"] = render
	funcMap["yield"] = func() (template.HTML, error) { return render(templateName) }

	if l := tmpl.render.Config.Layout; l != "" {
		tmpl.layout = l
	}

	if tmpl.layout != "" {
		p := filepath.Join("layouts", tmpl.layout)
		content, err = tmpl.findTemplate(p)
		if err != nil {
			return "findLayoutERR", errors.Wrapf(err, "render: Failed to find layout: %s", p)
		}

		t, err = template.New("").Funcs(funcMap).Parse(string(content))
		if err != nil {
			return "parseLayoutERR", errors.Wrapf(err, "render: Failed to parse layout: %s", p)
		}
		var b bytes.Buffer
		err = t.Execute(&b, obj)
		if err != nil {
			return "executeLayoutERR", errors.Wrapf(err, "render: failed to execute layout: %s", p)
		}
		return template.HTML(b.String()), nil
	}

	content, err = tmpl.findTemplate(templateName)
	if err != nil {
		return "findERR", errors.Wrapf(err, "failed to find template: %v", templateName)
	}
	t, err = template.New("").Funcs(funcMap).Parse(string(content))
	if err != nil {
		return "parseERROR", errors.Wrapf(err, "render: failed to parse template:%s", templateName)
	}
	var tpl bytes.Buffer
	err = t.Execute(&tpl, obj)
	if err != nil {
		return "executeERROR", errors.Wrapf(err, "render: failed to execute template:%s", templateName)
	}
	return template.HTML(tpl.String()), nil
}

// Execute execute tmpl
func (tmpl *Template) Execute(templateName string, obj interface{}, req *http.Request, w http.ResponseWriter) error {
	result, err := tmpl.Render(templateName, obj, req, w)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		// TODO: better error 500 template
		fmt.Fprint(w, "Internal Server Error")
		log := logging.FromContext(req.Context())
		log.Log("event", "error", "where", "tmpl.Execute", "err", err)
		return err
	}

	if w.Header().Get("Content-Type") == "" {
		w.Header().Set("Content-Type", "text/html")
	}

	_, err = io.Copy(w, bytes.NewReader([]byte(result)))
	return err
}

func (tmpl *Template) findTemplate(name string) ([]byte, error) {
	return tmpl.render.Asset(name + ".tmpl")
}
