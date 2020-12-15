package main

import (
	"eth2-exporter/utils"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"strings"

	"github.com/evanw/esbuild/pkg/api"
)

func bundle(staticDir string) error {
	if staticDir == "" {
		staticDir = "./static"
	}

	fileInfo, err := os.Stat(staticDir)
	if err != nil {
		return fmt.Errorf("error getting stats about the static dir", err)
	}

	if !fileInfo.IsDir() {
		return fmt.Errorf("error static dir is not a directory")
	}

	bundleDir := path.Join(staticDir, "bundle")
	if _, err := os.Stat(bundleDir); os.IsNotExist(err) {
		os.Mkdir(bundleDir, 0755)
	} else if err != nil {
		return fmt.Errorf("error getting stats about the bundle dir", err)
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
			return err
		}

		for _, match := range matches {
			code, err := ioutil.ReadFile(match)
			if err != nil {
				return fmt.Errorf("error reading file %v", err)
			}
			if !strings.Contains(match, ".min") {
				content := string(code)
				result := api.Transform(content, fileType.transform)
				if len(result.Errors) != 0 {
					return fmt.Errorf("error transforming %v %v", fileType, result.Errors)
				}
				code = result.Code
			}
			match = strings.Replace(match, typeDir, bundleTypeDir, -1)

			if _, err := os.Stat(path.Dir(match)); os.IsNotExist(err) {
				os.Mkdir(path.Dir(match), 0755)
			}

			err = ioutil.WriteFile(match, code, 0755)
			if err != nil {
				return fmt.Errorf("error failed to write file %v", err)
			}
		}
	}

	return nil
}

func main() {
	if err := bundle("./static"); err != nil {
		log.Fatal("error bundling: ", err)
	}
}
