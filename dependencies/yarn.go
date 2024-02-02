package dependencies

import (
	"bufio"
	"fmt"
	"io"
	"strings"

	"golang.org/x/exp/maps"
)

const YARN_DEPENDENCY_LINE_SUFFIX = ":"
const YARN_REMOTE_PREFIX = "  resolved \""
const YARN_DEP_VERSION_SEPARATOR = "@"

type YarnLock struct {
	Remotes map[string]Remote
}

func NewYarnLock() *YarnLock {
	return &YarnLock{
		Remotes: make(map[string]Remote),
	}
}

func (y *YarnLock) ParseLockFile(file io.Reader) {
	scanner := bufio.NewScanner(file)
	var currentDependencyName string
	readingDep := false

	for scanner.Scan() {
		line := scanner.Text()

		switch {
		case !readingDep && strings.HasSuffix(line, YARN_DEPENDENCY_LINE_SUFFIX):
			currentDependencyName = strings.TrimPrefix(line, "\"")

			if currentDependencyName[0] == '@' {
				currentDependencyName = "@" + strings.SplitN(currentDependencyName[1:], YARN_DEP_VERSION_SEPARATOR, 2)[0]
			} else {
				currentDependencyName = strings.SplitN(currentDependencyName, YARN_DEP_VERSION_SEPARATOR, 2)[0]
			}

			readingDep = true
		case readingDep && strings.HasPrefix(line, YARN_REMOTE_PREFIX):
			remoteStr := strings.TrimPrefix(line, YARN_REMOTE_PREFIX)
			remoteUrl := strings.SplitN(remoteStr, "/"+currentDependencyName+"/", 2)[0]

			_, ok := y.Remotes[remoteUrl]

			if !ok {
				y.addRemote(remoteUrl)
			}

			y.addDependency(remoteUrl, currentDependencyName)
		case line == "":
			readingDep = false
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Println("Error reading file:", err)
	}
}

func (y *YarnLock) addRemote(url string) {
	y.Remotes[url] = Remote{Url: url, Dependencies: make(map[string]Dependency)}
}

func (y *YarnLock) addDependency(remoteUrl, dependencyName string) {
	remote := y.Remotes[remoteUrl]
	remote.Dependencies[dependencyName] = Dependency{Name: dependencyName}
	y.Remotes[remoteUrl] = remote
}

func (y *YarnLock) PrintRemotes() {
	for remoteName, remote := range y.Remotes {
		fmt.Printf("%+v\n", remoteName)
		fmt.Printf("%+v\n", remote)
		fmt.Println("--------------------")
	}
}

func (g *YarnLock) GetRemotesUrlsWithDependencyMismatch(registryUrl, dependency string) []string {
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

func (y *YarnLock) MatchRemoteUrls(grep string) []string {
	if grep == "" {
		return maps.Keys(y.Remotes)
	}

	remoteUrls := make([]string, 0, len(y.Remotes))

	for url := range y.Remotes {
		if strings.Contains(strings.ToLower(url), strings.ToLower(grep)) {
			remoteUrls = append(remoteUrls, url)
		}
	}

	return remoteUrls
}
