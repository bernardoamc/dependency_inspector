package dependencies

import (
	"bufio"
	"fmt"
	"io"
	"regexp"
	"strings"

	"golang.org/x/exp/maps"
)

// Define constants for indentation to help our parser
//
// 1. Remote and spec are preceded by two spaces
// 2. Dependency is preceded by four spaces
// 3. Dependency is preceded by six spaces
//
// Example:
//
//	remote: https://rubygems.org/
//	specs:
//	  actioncable (5.2.2)
//	    actionpack (= 5.2.2)
const (
	GEMFILE_REMOTE_PREFIX     = "  remote:" // Two spaces for remote
	GEMFILE_SPEC_PREFIX       = "  specs:"  // Two spaces for spec
	GEMFILE_DEPENDENCY_REGEXP = `^\s{4}\S`  // Four spaces followed by a non-space character for dependency
)

type GemfileLock struct {
	Remotes map[string]Remote
}

func NewGemfileLock() *GemfileLock {
	return &GemfileLock{
		Remotes: make(map[string]Remote),
	}
}

func (g *GemfileLock) ParseLockFile(file io.Reader) {
	scanner := bufio.NewScanner(file)

	var currentRemoteName, currentDependencyName string
	readingRemote, readingSpecs := false, false
	dependencyRegexp := regexp.MustCompile(GEMFILE_DEPENDENCY_REGEXP)

	for scanner.Scan() {
		line := scanner.Text()

		switch {
		case strings.HasPrefix(line, GEMFILE_REMOTE_PREFIX):
			currentRemoteName = strings.TrimSpace(strings.TrimPrefix(line, GEMFILE_REMOTE_PREFIX))
			g.addRemote(currentRemoteName)
			readingRemote = true
		case readingRemote && strings.HasPrefix(line, GEMFILE_SPEC_PREFIX):
			readingSpecs = true
		case readingSpecs && dependencyRegexp.MatchString(line):
			currentDependencyName = strings.Fields(line)[0]
			g.addDependency(currentRemoteName, currentDependencyName)
		case line == "":
			readingRemote = false
			readingSpecs = false
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Println("Error reading file:", err)
	}
}

func (g *GemfileLock) addRemote(url string) {
	g.Remotes[url] = Remote{Url: url, Dependencies: make(map[string]Dependency)}
}

func (g *GemfileLock) addDependency(remoteUrl, dependencyName string) {
	remote := g.Remotes[remoteUrl]
	remote.Dependencies[dependencyName] = Dependency{Name: dependencyName}
	g.Remotes[remoteUrl] = remote
}

func (g *GemfileLock) GetRemotes() map[string]Remote {
	return g.Remotes
}

func (g *GemfileLock) PrintRemotes() {
	for remoteName, remote := range g.Remotes {
		fmt.Printf("%+v\n", remoteName)
		fmt.Printf("%+v\n", remote)
		fmt.Println("--------------------")
	}
}

func (g *GemfileLock) GetRemotesUrlsWithDependencyMismatch(registryUrl, dependency string) []string {
	remotes := make([]string, 0)

	for url, remote := range g.Remotes {
		if registryUrl == remote.Url {
			continue
		}

		if remote.HasDependency(dependency) {
			remotes = append(remotes, url)
		}
	}

	return remotes
}

func (g *GemfileLock) MatchRemoteUrls(grep string) []string {
	if grep == "" {
		return maps.Keys(g.Remotes)
	}

	remoteUrls := make([]string, 0, len(g.Remotes))

	for url := range g.Remotes {
		if strings.Contains(strings.ToLower(url), strings.ToLower(grep)) {
			remoteUrls = append(remoteUrls, url)
		}
	}

	return remoteUrls
}
