package utils

import (
	"bytes"
	"html/template"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"

	zglob "github.com/mattn/go-zglob"
	"github.com/palantir/stacktrace"
)

// Version of rivendell
var Version = "1.0.0"

// StringSet .
type StringSet map[string]struct{}

// NewStringSet .
func NewStringSet(values ...string) StringSet {
	s := make(StringSet)
	s.Add(values...)
	return s
}

// Add new values to string set
func (s StringSet) Add(values ...string) {
	for _, value := range values {
		s[value] = struct{}{}
	}
}

// Remove values from string set
func (s StringSet) Remove(values ...string) {
	for _, value := range values {
		delete(s, value)
	}
}

// Exists checks a value appears in the set or not
func (s StringSet) Exists(value string) bool {
	_, ok := s[value]
	return ok
}

// ToSlice converts the string set to a slice
func (s StringSet) ToSlice() []string {
	slice := []string{}
	for value := range s {
		slice = append(slice, value)
	}
	return slice
}

// First returns first element from the set
func (s StringSet) First() string {
	for value := range s {
		return value
	}
	return ""
}

// MergeMaps merges multiple maps into one
func MergeMaps(maps ...map[string]string) map[string]string {
	finalMap := make(map[string]string)
	for _, m := range maps {
		for k, v := range m {
			finalMap[k] = v
		}
	}
	return finalMap
}

// ExecuteTemplate .
func ExecuteTemplate(templateFile string, variables map[string]string) ([]byte, error) {
	content, err := ioutil.ReadFile(templateFile)
	if err != nil {
		return nil, stacktrace.Propagate(err, "cannot read template file %q", templateFile)
	}
	contentWithEnvExpand := os.ExpandEnv(string(content))
	tmpl, err := template.New("rivendell").Parse(contentWithEnvExpand)
	if err != nil {
		return nil, stacktrace.Propagate(err, "cannot parse template file %q", templateFile)
	}
	tmpl = tmpl.Option("missingkey=error")
	b := &bytes.Buffer{}
	err = tmpl.Execute(b, variables)
	if err != nil {
		return nil, stacktrace.Propagate(err, "cannot execute template file %q", templateFile)
	}
	return b.Bytes(), nil
}

// GlobFiles returns list of files based on inclusion and exclusion pattern
func GlobFiles(includePatterns, excludePatterns []string) ([]string, error) {
	candidates := make(StringSet)
	for _, includePattern := range includePatterns {
		files, err := zglob.Glob(includePattern)
		if err != nil {
			return nil, stacktrace.Propagate(err, "cannot find files from pattern %q", includePattern)
		}
		candidates.Add(files...)
	}
	for _, excludePattern := range excludePatterns {
		excludedFiles, err := zglob.Glob(excludePattern)
		if err != nil {
			return nil, stacktrace.Propagate(err, "cannot find files from pattern %q", excludePattern)
		}
		candidates.Remove(excludedFiles...)
	}
	slice := candidates.ToSlice()
	sort.Strings(slice)
	return slice, nil
}

// NilArrayToEmpty converts a nil array to empty array if possible
func NilArrayToEmpty(a []string) []string {
	if a == nil {
		return []string{}
	}
	return a
}

// StringArrayMap .
func StringArrayMap(a []string, f func(string) string) []string {
	result := []string{}
	for _, x := range a {
		result = append(result, f(x))
	}
	return result
}

// PrependPaths prepends a prefix to an array of paths
func PrependPaths(prefix string, paths []string) []string {
	return StringArrayMap(paths, func(path string) string {
		return filepath.Join(prefix, path)
	})
}
