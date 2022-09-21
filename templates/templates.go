package templates

import (
	"embed"
	"eth2-exporter/utils"
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"
	"sync"
	"text/template"

	"github.com/sirupsen/logrus"
)

var (
	//go:embed *
	Files embed.FS
)

var templateCache = make(map[string]*template.Template)
var templateCacheMux = &sync.RWMutex{}
var templateFuncs = utils.GetTemplateFuncs()
var pages map[string][]string

// compile time check for templates
var _ error = CompileTimeCheck(fs.FS(Files))

func GetTemplate(name string, files ...string) *template.Template {
	if utils.Config.Frontend.Debug {
		for i := range files {
			files[i] = "templates/" + files[i]
		}
		return template.Must(template.New(name).Funcs(template.FuncMap(templateFuncs)).ParseFiles(files...))
	}

	name = name + strings.Join(files, "-")

	templateCacheMux.RLock()
	defer templateCacheMux.RUnlock()
	if templateCache[name] != nil {
		return templateCache[name]
	}

	tmpl := template.Must(template.New(name).Funcs(template.FuncMap(templateFuncs)).ParseFS(Files, files...))
	templateCacheMux.Lock()
	templateCache[name] = tmpl
	templateCacheMux.Unlock()

	return templateCache[name]
}

func CompileTimeCheck(fsys fs.FS) error {
	files, err := getFileSysNames(fsys, ".")
	if err != nil {
		return err
	}
	template.Must(template.New("layout").Funcs(template.FuncMap(templateFuncs)).ParseFS(Files, files...))
	logrus.Infof("Compile Time Check Complete")

	return nil
}

func getFileSysNames(fsys fs.FS, dirname string) ([]string, error) {
	entry, err := fs.ReadDir(fsys, dirname)
	if err != nil {
		return nil, fmt.Errorf("error reading embed directory, err: %w", err)
	}

	files := make([]string, 0, 100)
	for _, f := range entry {
		info, err := f.Info()
		if err != nil {
			return nil, fmt.Errorf("error returning file info err: %w", err)
		}
		if !f.IsDir() {
			files = append(files, filepath.Join(dirname, info.Name()))
		} else {
			names, err := getFileSysNames(fsys, info.Name())
			if err != nil {
				return nil, err
			}
			files = append(files, names...)
		}
	}

	return files, nil
}
