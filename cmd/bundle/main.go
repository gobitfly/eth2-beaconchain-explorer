package main

import (
	"crypto/md5"
	"eth2-exporter/utils"
	"fmt"
	"log"
	"os"
	"path"
	"strings"

	"github.com/evanw/esbuild/pkg/api"
)

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
		os.Mkdir(bundleDir, 0755)
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
			code, err := os.ReadFile(match)
			if err != nil {
				return nameMapping, fmt.Errorf("error reading file %v", err)
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
				os.Mkdir(path.Dir(matchBundle), 0755)
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
	files, err := bundle("./static")
	if err != nil {
		log.Fatalf("error bundling: %v", err)
	}

	if err := replaceFilesNames(files); err != nil {
		log.Fatalf("error replacing dependencies err: %v", err)
	}
}
