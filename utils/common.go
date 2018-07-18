package utils

import (
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	zglob "github.com/mattn/go-zglob"
	"github.com/palantir/stacktrace"
)

// Version of rivendell
var Version = "1.0.7"

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

var envVarRegex = regexp.MustCompile("\\$\\([^\\)]+\\)")

// ExpandEnv replaces $(var) with environment variable value
func ExpandEnv(s string) string {
	return envVarRegex.ReplaceAllStringFunc(s, func(found string) string {
		envName := strings.TrimPrefix(strings.TrimSuffix(found, ")"), "$(")
		return os.Getenv(envName)
	})
}
