package utils

import (
	"bufio"
	"bytes"
	"io/ioutil"
	"path/filepath"
	"strings"
	"text/template"

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
	tmpl, err := template.New(templateFile).Funcs(map[string]interface{}{
		"import": importFunc,
		"indent": indentFunc,
	}).Parse(contentWithEnvExpand)
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
	var realPath string
	if strings.HasPrefix(templateFile, "/") {
		realPath = templateFile
	} else {
		currentFolder, err := cwdStack.Head()
		if err != nil {
			return "", err
		}
		realPath = filepath.Join(currentFolder, templateFile)
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
