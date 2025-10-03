package templates

import (
	"embed"
	"fmt"
	"html/template"
	"io/fs"
	"path/filepath"
	"strings"
	"sync"

	"github.com/gobitfly/eth2-beaconchain-explorer/utils"

	"github.com/sirupsen/logrus"
)

var logger = logrus.New().WithField("module", "templates")

var (
	//go:embed *
	Files embed.FS
)

var templateCache = make(map[string]*template.Template)
var templateCacheMux = &sync.RWMutex{}
var templateFuncs = utils.GetTemplateFuncs()

func GetTemplate(files ...string) *template.Template {
	name := strings.Join(files, "-")

	if utils.Config.Frontend.Debug {
		templateFiles := make([]string, 0, len(files))
		for _, file := range files {
			if strings.Contains(file, "*") {
				if !strings.HasPrefix(file, "templates") {
					file = "templates/" + file
				}
				matches, err := filepath.Glob(file)
				if err != nil {
					logger.Errorf("error globbing template files: %s", err)
					continue
				}
				templateFiles = append(templateFiles, matches...)
			} else if strings.HasPrefix(file, "templates") {
				templateFiles = append(templateFiles, file)
			} else {
				templateFiles = append(templateFiles, "templates/"+file)
			}
		}
		return template.Must(template.New(name).Funcs(templateFuncs).ParseFiles(templateFiles...))
	}

	templateCacheMux.RLock()
	if templateCache[name] != nil {
		defer templateCacheMux.RUnlock()
		return templateCache[name]
	}
	templateCacheMux.RUnlock()

	tmpl := template.Must(template.New(name).Funcs(templateFuncs).ParseFS(Files, files...))
	templateCacheMux.Lock()
	defer templateCacheMux.Unlock()
	templateCache[name] = tmpl
	return templateCache[name]
}

func GetTemplateNames() []string {
	files, _ := getFileSysNames(fs.FS(Files), ".")
	return files
}

func CompileTimeCheck(fsys fs.FS) error {
	files, err := getFileSysNames(fsys, ".")
	if err != nil {
		return err
	}
	template.Must(template.New("layout").Funcs(template.FuncMap(templateFuncs)).ParseFS(Files, files...))
	logger.Infof("compile time check completed")

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
