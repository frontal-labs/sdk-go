// Package templates provides Go starter projects for common Frontal SDK setups.
package templates

import (
	"embed"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"text/template"
)

//go:embed cli/*.tmpl http-service/*.tmpl worker/*.tmpl stream-consumer/*.tmpl
var files embed.FS

var modulePathPattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._-]*(/[A-Za-z0-9][A-Za-z0-9._-]*)+$`)

type templateData struct {
	Module     string
	SDKReplace string
}

// Available returns the names of all starter templates.
func Available() []string {
	names := []string{"cli", "http-service", "stream-consumer", "worker"}
	sort.Strings(names)
	return names
}

// Render writes a starter project to a new directory. sdkPath is the relative
// path from destination to the SDK checkout used by the generated go.mod.
func Render(name, module, destination, sdkPath string) error {
	if !contains(Available(), name) {
		return fmt.Errorf("unknown template %q", name)
	}
	if !modulePathPattern.MatchString(module) {
		return fmt.Errorf("invalid Go module path %q", module)
	}
	if sdkPath == "" || filepath.IsAbs(sdkPath) {
		return errors.New("SDK path must be a non-empty relative path")
	}
	destination, err := filepath.Abs(destination)
	if err != nil {
		return fmt.Errorf("resolve destination: %w", err)
	}
	if _, err := os.Stat(destination); err == nil {
		return fmt.Errorf("destination %q already exists", destination)
	} else if !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("inspect destination: %w", err)
	}
	parent := filepath.Dir(destination)
	if err := os.MkdirAll(parent, 0o755); err != nil {
		return fmt.Errorf("create destination parent: %w", err)
	}
	staging, err := os.MkdirTemp(parent, ".frontal-template-*")
	if err != nil {
		return fmt.Errorf("create staging directory: %w", err)
	}
	defer func() { _ = os.RemoveAll(staging) }()

	data := templateData{Module: module, SDKReplace: filepath.ToSlash(sdkPath)}
	for _, filename := range []string{"go.mod.tmpl", "main.go.tmpl", "README.md.tmpl"} {
		sourcePath := filepath.ToSlash(filepath.Join(name, filename))
		content, err := fs.ReadFile(files, sourcePath)
		if err != nil {
			return fmt.Errorf("read %s template: %w", filename, err)
		}
		parsed, err := template.New(filename).Funcs(template.FuncMap{
			"quote": func(value string) string { return fmt.Sprintf("%q", value) },
		}).Parse(string(content))
		if err != nil {
			return fmt.Errorf("parse %s template: %w", filename, err)
		}
		outputName := strings.TrimSuffix(filename, ".tmpl")
		output, err := os.OpenFile(filepath.Join(staging, outputName), os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o644)
		if err != nil {
			return fmt.Errorf("create %s: %w", outputName, err)
		}
		executeErr := parsed.Execute(output, data)
		closeErr := output.Close()
		if executeErr != nil {
			return fmt.Errorf("render %s: %w", outputName, executeErr)
		}
		if closeErr != nil {
			return fmt.Errorf("close %s: %w", outputName, closeErr)
		}
	}
	if err := os.Rename(staging, destination); err != nil {
		return fmt.Errorf("install generated project: %w", err)
	}
	return nil
}

func contains(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}
