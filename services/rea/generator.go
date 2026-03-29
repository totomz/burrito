package rea

import (
	"bytes"
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"path"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/totomz/burrito/common"
)

type ServiceSeed struct {
	ServiceName string
}

func (service *ServiceSeed) ServicePath() string {

	wd, err := projectRoot(common.MustGetCwd())
	if err != nil {
		panic(err)
	}
	return path.Join(wd, "services", service.ServiceName)
}

// CreateServiceDirectory creates a new directory for the service.
// Returns an error if the directory cannot be created.
func CreateServiceDirectory(seed ServiceSeed) error {
	err := os.Mkdir(seed.ServicePath(), 0755)
	if err != nil {
		return fmt.Errorf("error creating service directory: %w", err)
	}

	return nil
}

// DeleteServiceDirectory deletes the service directory and all its contents.
// Returns an error if the directory cannot be deleted.
func DeleteServiceDirectory(seed ServiceSeed) error {
	err := os.RemoveAll(seed.ServicePath())
	if err != nil {
		return fmt.Errorf("error deleting service directory: %w", err)
	}

	return nil
}

func projectRoot(path string) (string, error) {

	if filepath.Base(path) == "template-burrito" {
		return path, nil
	}

	if filepath.Base(path) == "." || filepath.Base(path) == "/" {
		return "", fmt.Errorf("project root template-burrito not found in path")
	}

	return projectRoot(filepath.Dir(path))
}

// CloneTemplate copies all Go template files from tpl_go_kube_v1 to the service path,
// rendering them with the ServiceSeed data. It retains the folder structure and creates
// folders that do not exist.
func CloneTemplate(seed ServiceSeed) error {
	wd, err := projectRoot(common.MustGetCwd())
	if err != nil {
		return err
	}

	templateDir := filepath.Join(wd, "services", "rea", "tpl_go_kube_v1")

	slog.Info("cloning template", "from", templateDir, "to", seed.ServicePath())

	// Walk through all files in the template directory
	err = filepath.WalkDir(templateDir, func(srcPath string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Get the relative path from the template directory
		relPath, err := filepath.Rel(templateDir, srcPath)
		if err != nil {
			return fmt.Errorf("error getting relative path: %w", err)
		}

		// Skip the root directory itself
		if relPath == "." {
			return nil
		}

		// Construct the destination path
		dstPath := filepath.Join(seed.ServicePath(), relPath)

		// If it's a directory, create it
		if d.IsDir() {
			err := os.MkdirAll(dstPath, 0755)
			if err != nil {
				return fmt.Errorf("error creating directory %s: %w", dstPath, err)
			}
			slog.Info("created directory", "path", dstPath)
			return nil
		}

		// If it's a file, read it, render it as a template, and write it
		fileContentRaw, err := os.ReadFile(srcPath)
		if err != nil {
			return fmt.Errorf("error reading file %s: %w", srcPath, err)
		}

		// go files are using the `rea` name so they compile and we can use syntax-highlighting
		// while working at the templates.
		fileContentTpl := strings.ReplaceAll(string(fileContentRaw), "rea", "[[.ServiceName]]")

		// Parse and execute the template
		// Use `[[ ]]` as delimitr, to handle the helm templates
		tmpl, err := template.New(filepath.Base(srcPath)).Delims("[[", "]]").Parse(fileContentTpl)
		if err != nil {
			return fmt.Errorf("error parsing template %s: %w", srcPath, err)
		}

		var renderedContent bytes.Buffer
		err = tmpl.Execute(&renderedContent, seed)
		if err != nil {
			return fmt.Errorf("error executing template %s: %w", srcPath, err)
		}

		// Write the rendered content to the destination
		err = os.WriteFile(dstPath, renderedContent.Bytes(), 0644)
		if err != nil {
			return fmt.Errorf("error writing file %s: %w", dstPath, err)
		}

		slog.Info("created file", "path", dstPath)
		return nil
	})
	if err != nil {
		return fmt.Errorf("error walking template directory: %w", err)
	}

	// Rename the folder with the helm chart
	oldChart := path.Join(seed.ServicePath(), "rea")
	newChart := path.Join(seed.ServicePath(), seed.ServiceName)
	err = os.Rename(oldChart, newChart)
	if err != nil {
		return fmt.Errorf("error renaming chart folder: %w", err)
	}

	slog.Info("Done!")
	println("######")
	println("Add `source ./services/zalice/shMakefile` to the main shMakefile!")
	println("######")
	return nil
}
