package project

import (
	"bufio"
	"bytes"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/anduintransaction/rivendell/utils"
	"github.com/palantir/stacktrace"
	yaml "gopkg.in/yaml.v2"
)

type ResourceFileProcessor interface {
	Process(f *ResourceFile) error
}

type ResourceFileProcessorFunc func(*ResourceFile) error

func (f ResourceFileProcessorFunc) Process(rf *ResourceFile) error {
	return f(rf)
}

// resolveResourcePath resolve `resources` config into a list of ResourceFile. Support 2 type of path:
//  - URL path: either start with `http` or `https`. Will fetch from internet
//  - File globs: paths that does not start with http
// modify ResourceGroup directly
// Note: this function does not expand resource file content with variables automatically
func resolveResourceFile(rootDir string, group *ResourceGroupConfig, globalInclude, globalExclude []string) ([]*ResourceFile, error) {
	urlPattern := utils.StringArrayFilter(group.Resources, utils.IsURL)
	globPattern := utils.StringArrayFilter(group.Resources, func(s string) bool {
		return !utils.IsURL(s)
	})

	globPattern = utils.PrependPaths(rootDir, globPattern)
	globalInclude = utils.PrependPaths(rootDir, globalInclude)
	groupExcludes := utils.PrependPaths(rootDir, append(group.Excludes, globalExclude...))
	globFiles, err := resolveResourceFileByGlob(rootDir, globPattern, groupExcludes, globalInclude)
	if err != nil {
		return nil, err
	}

	urlFiles, err := resolveResourceFileByURL(rootDir, urlPattern)
	if err != nil {
		return nil, err
	}

	return append(globFiles, urlFiles...), nil
}

func resolveResourceFileByGlob(rootDir string, include, exclude, join []string) ([]*ResourceFile, error) {
	ret := []*ResourceFile{}

	fPaths, err := utils.GlobFiles(include, exclude)
	if err != nil {
		return nil, err
	}
	if len(join) > 0 {
		joinPaths, err := utils.GlobFiles(join, nil)
		if err != nil {
			return nil, err
		}

		inclusiveSet := utils.NewStringSet(fPaths...)
		joinSet := utils.NewStringSet(joinPaths...)
		fPaths = inclusiveSet.Join(joinSet).ToSlice()
	}
	sort.Slice(fPaths, func(i, j int) bool {
		return fPaths[i] < fPaths[j]
	})

	for _, p := range fPaths {
		stat, err := os.Stat(p)
		if err != nil {
			return nil, err
		}
		if stat.IsDir() {
			continue
		}
		content, err := ioutil.ReadFile(p)
		if err != nil {
			return nil, err
		}

		ret = append(ret, &ResourceFile{
			Source:     p,
			ContextDir: filepath.Dir(p),
			RawContent: string(content),
		})
	}

	return ret, nil
}

func resolveResourceFileByURL(rootDir string, pattern []string) ([]*ResourceFile, error) {
	ret := []*ResourceFile{}

	for _, url := range pattern {
		resp, err := http.Get(url)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		content, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}

		ret = append(ret, &ResourceFile{
			Source:     url,
			ContextDir: rootDir,
			RawContent: string(content),
		})
	}

	return ret, nil
}

func expandResourceContent(variables map[string]string) ResourceFileProcessorFunc {
	return func(rf *ResourceFile) error {
		expandedContent, err := utils.ExecuteTemplateContent(
			rf.ContextDir,
			[]byte(rf.RawContent),
			variables,
		)
		if err != nil {
			return err
		}

		rf.ExpandedContent = string(expandedContent)
		return nil
	}
}

func stripNamespace() ResourceFileProcessorFunc {
	return func(rf *ResourceFile) error {
		scanner := bufio.NewScanner(bytes.NewBuffer([]byte(rf.ExpandedContent)))
		strippedContent := ""
		for scanner.Scan() {
			line := scanner.Text()
			if !strings.HasPrefix(line, "  namespace:") {
				strippedContent += line + "\n"
			}
		}
		rf.ExpandedContent = strippedContent
		return nil
	}
}

func splitResourceContent() ResourceFileProcessorFunc {
	return func(rf *ResourceFile) error {
		parts := strings.Split(string(rf.ExpandedContent), "---\n")
		for _, part := range parts {
			part = strings.TrimSpace(part)
			if len(part) == 0 {
				continue
			}
			parsedResource := &resourceYAML{}
			err := yaml.Unmarshal([]byte(part), parsedResource)
			if err != nil {
				return stacktrace.Propagate(err, "Cannot parse yaml file %q. Content: %s", rf.Source, part)
			}
			resource := &Resource{
				Name:       parsedResource.Metadata.Name,
				Kind:       parsedResource.Kind,
				Filepath:   rf.Source,
				RawContent: part,
			}
			rf.Resources = append(rf.Resources, resource)
		}
		return nil
	}
}
