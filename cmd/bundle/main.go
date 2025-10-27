package main

import (
	"crypto/md5"
	"flag"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/gobitfly/eth2-beaconchain-explorer/utils"

	"github.com/evanw/esbuild/pkg/api"
)

var tsSourceMap = flag.Bool("ts-sourcemap", false, "emit inline sourcemaps for TS (dev)")

// buildTypeScript compiles all TS/TSX under static/ into static/js/[name].js
func buildTypeScript(staticDir string) error {
	// Only explicit entry files; imports will be bundled into those outputs.
	isEntry := func(p string) bool {
		return strings.HasSuffix(p, ".entry.ts") || strings.HasSuffix(p, ".entry.tsx")
	}

	var entries []string
	err := filepath.WalkDir(staticDir, func(p string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() {
			switch d.Name() {
			case "js", "bundle", "node_modules":
				return filepath.SkipDir
			}
			return nil
		}
		if strings.HasSuffix(p, ".d.ts") {
			return nil
		}
		if (strings.HasSuffix(p, ".ts") || strings.HasSuffix(p, ".tsx")) && isEntry(p) {
			entries = append(entries, p)
		}
		return nil
	})
	if err != nil {
		return err
	}
	if len(entries) == 0 {
		return nil
	}

	opts := api.BuildOptions{
		EntryPoints: entries,
		Outdir:      path.Join(staticDir, "js"),
		Outbase:     staticDir,
		Bundle:      true,
		Format:      api.FormatESModule,
		Platform:    api.PlatformBrowser,
		Loader: map[string]api.Loader{
			".ts":   api.LoaderTS,
			".tsx":  api.LoaderTSX,
			".json": api.LoaderJSON,
		},
		// Add source maps (inline for dev only)
		Sourcemap: func() api.SourceMap {
				if tsSourceMap != nil && *tsSourceMap {
						return api.SourceMapInline
				}
				return api.SourceMapNone
		}(),
		Write:    true,
		LogLevel: api.LogLevelInfo,
	}

	result := api.Build(opts)
	if len(result.Errors) > 0 {
		return fmt.Errorf("ts build failed: %v", result.Errors)
	}

	return nil
}

// Very small watcher for .ts/.tsx that calls buildTypeScript once per change.
func watchTypeScript(staticDir string) error {
	// initial build
	if err := buildTypeScript(staticDir); err != nil {
		return err
	}

	w, err := fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("watcher init: %w", err)
	}
	defer w.Close()

	// watch all subdirs under static/, except outputs to avoid loops
	err = filepath.WalkDir(staticDir, func(p string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return nil
		}
		if d.IsDir() {
			base := filepath.Base(p)
			switch base {
			case "js", "bundle", "node_modules":
				return filepath.SkipDir
			}
			_ = w.Add(p)
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("walk watch dirs: %w", err)
	}

	isOutput := func(p string) bool {
		sep := string(filepath.Separator)
		return strings.Contains(p, sep+"js"+sep) || strings.Contains(p, sep+"bundle"+sep)
	}
	okExt := func(p string) bool {
		ext := strings.ToLower(filepath.Ext(p))
		return ext == ".ts" || ext == ".tsx"
	}

	// debounce rapid events
	var timer *time.Timer
	trigger := func() {
		if timer != nil {
			timer.Stop()
		}
		timer = time.AfterFunc(200*time.Millisecond, func() {
			if err := buildTypeScript(staticDir); err != nil {
				log.Printf("TS rebuild failed: %v", err)
			} else {
				log.Println("TS rebuilt")
			}
		})
	}

	log.Println("Watching TypeScript for changes...")
	for {
		select {
		case ev := <-w.Events:
			if ev.Op&(fsnotify.Create|fsnotify.Write|fsnotify.Rename|fsnotify.Remove) == 0 {
				continue
			}
			if isOutput(ev.Name) || !okExt(ev.Name) {
				continue
			}

			if ev.Op&fsnotify.Create != 0 {
				if fi, e := os.Stat(ev.Name); e == nil && fi.IsDir() {
					_ = w.Add(ev.Name)
				}
			}
			trigger()

		case e := <-w.Errors:
			log.Printf("watch error: %v", e)
		}
	}
}

func bundle(staticDir string) (map[string]string, error) {

	nameMapping := make(map[string]string, 0)

	if staticDir == "" {
		staticDir = "./static"
	}

	fileInfo, err := os.Stat(staticDir)
	if err != nil {
		return nameMapping, fmt.Errorf("error getting stats about the static dir: %v", err)
	}

	if !fileInfo.IsDir() {
		return nameMapping, fmt.Errorf("error static dir is not a directory")
	}

	bundleDir := path.Join(staticDir, "bundle")
	if _, err := os.Stat(bundleDir); os.IsNotExist(err) {
		os.MkdirAll(bundleDir, 0755)
	} else if err != nil {
		return nameMapping, fmt.Errorf("error getting stats about the bundle dir: %v", err)
	}

	type fileType struct {
		ext       string
		transform api.TransformOptions
	}

	types := []fileType{
		{
			ext: "css",
			transform: api.TransformOptions{
				Loader:            api.LoaderCSS,
				MinifyWhitespace:  true,
				MinifyIdentifiers: false,
				MinifySyntax:      true,
			},
		},
		{
			ext: "js",
			transform: api.TransformOptions{
				Loader:            api.LoaderJS,
				MinifyWhitespace:  true,
				MinifyIdentifiers: false,
				MinifySyntax:      true,
			},
		},
	}

	for _, fileType := range types {
		bundleTypeDir := path.Join(bundleDir, fileType.ext)
		typeDir := path.Join(staticDir, fileType.ext)
		matches, err := utils.Glob(typeDir, "."+fileType.ext)
		if err != nil {
			return nameMapping, err
		}

		for _, match := range matches {
			code, readErr := os.ReadFile(match)
			if readErr != nil {
				return nameMapping, fmt.Errorf("error reading file %v", readErr)
			}
			if !strings.Contains(match, ".min") {
				content := string(code)
				result := api.Transform(content, fileType.transform)
				if len(result.Errors) != 0 {
					return nameMapping, fmt.Errorf("error transforming %v %v", fileType, result.Errors)
				}
				code = result.Code
			}
			matchBundle := strings.Replace(match, typeDir, bundleTypeDir, -1)

			if _, err := os.Stat(path.Dir(matchBundle)); os.IsNotExist(err) {
				os.MkdirAll(path.Dir(matchBundle), 0755)
			}

			codeHash := fmt.Sprintf("%x", md5.Sum([]byte(code)))
			matchHash := strings.Replace(matchBundle, "."+fileType.ext, "."+codeHash[:6]+"."+fileType.ext, -1)

			path := strings.ReplaceAll(match, "static/", "")
			newPath := strings.ReplaceAll(matchHash, "static/", "")
			nameMapping[path] = newPath

			err = os.WriteFile(matchHash, code, 0755)
			if err != nil {
				return nameMapping, fmt.Errorf("error failed to write file %v", err)
			}
		}
	}

	return nameMapping, nil
}

func replaceFilesNames(files map[string]string) error {
	templates := "./bin/templates"
	templatesDir := path.Join(templates)

	matches, err := utils.Glob(templatesDir, ".html")
	if err != nil {
		return err
	}
	for _, match := range matches {
		html, err := os.ReadFile(match)
		if err != nil {
			return err
		}
		h := string(html)
		for oldPath, newPath := range files {
			// logrus.Info("replacing: ", oldPath, " with: ", newPath)
			h = strings.ReplaceAll(h, oldPath, newPath)
		}
		err = os.WriteFile(match, []byte(h), 0755)
		if err != nil {
			return err
		}
	}

	return nil
}

func main() {
    staticDir := flag.String("static", "./static", "path to static directory")
    watchTS := flag.Bool("watch-ts", false, "watch and rebuild TypeScript on changes (dev only)")
    compileTS := flag.Bool("compile-ts", false, "compile TypeScript assets before bundling (dev only)")
    flag.Parse()

    if *watchTS {
        if err := watchTypeScript(*staticDir); err != nil {
            log.Fatal(err)
        }
        return
    }

    if *compileTS {
        if err := buildTypeScript(*staticDir); err != nil {
            log.Fatalf("error compiling typescript: %v", err)
        }
    }

    files, err := bundle(*staticDir)
    if err != nil {
        log.Fatalf("error bundling: %v", err)
    }

    if err := replaceFilesNames(files); err != nil {
        log.Fatalf("error replacing dependencies err: %v", err)
    }
}
