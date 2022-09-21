package templates

import (
	"embed"
)

var (
	//go:embed *
	Files embed.FS
)

// var templateCache = make(map[string]*template.Template)
// var templateCacheMux = &sync.RWMutex{}

// func GetTemplate(name string, files ...string) *template.Template {
// 	if utils.Config.Frontend.Debug {
// 		for i := range files {
// 			files[i] = "templates/" + files[i]
// 		}
// 		return template.Must(template.New(name).Funcs(utils.GetTemplateFuncs()).ParseFiles(files...))
// 	}

// 	name = name + strings.Join(files, "-")

// 	templateCacheMux.RLock()
// 	defer templateCacheMux.RUnlock()
// 	if templateCache[name] != nil {
// 		return templateCache[name]
// 	}

// 	tmpl := template.Must(template.New(name).Funcs(utils.GetTemplateFuncs()).ParseFS(templates.Files, files...))
// 	templateCacheMux.Lock()
// 	templateCache[name] = tmpl
// 	templateCacheMux.Unlock()

// 	return templateCache[name]
// }
