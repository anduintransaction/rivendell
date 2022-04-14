package utils

import (
	"bufio"
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/Masterminds/sprig/v3"
	"github.com/palantir/stacktrace"
)

var cwdStack = NewStringStack()

// ExecuteTemplate .
func ExecuteTemplate(templateFile string, variables map[string]string) ([]byte, error) {
	content, err := ioutil.ReadFile(templateFile)
	if err != nil {
		return nil, stacktrace.Propagate(err, "cannot read template file %q", templateFile)
	}
	currentFolder := filepath.Dir(templateFile)
	cwdStack.Push(currentFolder)
	contentWithEnvExpand := ExpandEnv(string(content))
	tmpl, err := template.
		New(templateFile).
		Funcs(sprig.TxtFuncMap()).
		Funcs(map[string]interface{}{
			"import":       importFunc,
			"indent":       indentFunc,
			"loadFile":     loadFileFunc,
			"trim":         trimFunc,
			"hash":         hashFunc,
			"base64":       base64Func,
			"asGenericMap": asGenericMap,
			"asMapString":  asMapString,
		}).
		Parse(contentWithEnvExpand)
	if err != nil {
		return nil, stacktrace.Propagate(err, "cannot parse template file %q", templateFile)
	}
	tmpl = tmpl.Option("missingkey=error")
	b := &bytes.Buffer{}
	err = tmpl.Execute(b, variables)
	if err != nil {
		return nil, stacktrace.Propagate(err, "cannot execute template file %q", templateFile)
	}
	cwdStack.Pop()
	return b.Bytes(), nil
}

func importFunc(templateFile string, variables map[string]string) (string, error) {
	realPath, err := resolveRealpath(templateFile)
	if err != nil {
		return "", err
	}
	content, err := ExecuteTemplate(realPath, variables)
	return string(content), err
}

func indentFunc(indent int, content string) string {
	indentStr := strings.Repeat(" ", indent)
	result := ""
	scanner := bufio.NewScanner(bytes.NewReader([]byte(content)))
	for scanner.Scan() {
		result += indentStr + scanner.Text() + "\n"
	}
	return result
}

func loadFileFunc(filename string) (string, error) {
	realPath, err := resolveRealpath(filename)
	if err != nil {
		return "", err
	}
	content, err := ioutil.ReadFile(realPath)
	if err != nil {
		return "", err
	}
	return string(content), nil
}

func resolveRealpath(filename string) (string, error) {
	var realPath string
	if strings.HasPrefix(filename, "/") {
		realPath = filename
	} else {
		currentFolder, err := cwdStack.Head()
		if err != nil {
			return "", err
		}
		realPath = filepath.Join(currentFolder, filename)
	}
	return realPath, nil
}

func trimFunc(content string) string {
	return strings.TrimSpace(content)
}

func hashFunc(filename string) (string, error) {
	realPath, err := resolveRealpath(filename)
	if err != nil {
		return "", err
	}
	content, err := ioutil.ReadFile(realPath)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", sha256.Sum256(content)), nil
}

func base64Func(content string) string {
	return base64.StdEncoding.EncodeToString([]byte(content))
}

func asGenericMap(m map[string]string) map[string]interface{} {
	ret := make(map[string]interface{}, len(m))
	for k, v := range m {
		ret[k] = v
	}
	return ret
}

func asMapString(m map[string]interface{}) map[string]string {
	ret := make(map[string]string)
	for k, v := range m {
		if strVal, ok := v.(string); ok {
			ret[k] = strVal
		}
		if strg, ok := v.(fmt.Stringer); ok {
			ret[k] = strg.String()
		}
	}
	return ret
}
