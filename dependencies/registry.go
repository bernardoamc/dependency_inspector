package dependencies

import (
	"dependency_inspector/interfaces"
	"encoding/json"
	"io"
	"os"
)

type Registry struct {
	Url          string
	Dependencies []string
}

func BuildRegistry(path string) (Registry, error) {
	registry := Registry{
		Url:          "",
		Dependencies: make([]string, 0),
	}

	registryFile, err := os.Open(path)
	if err != nil {
		return Registry{}, err
	}
	defer registryFile.Close()

	registryContent, _ := io.ReadAll(registryFile)
	err = json.Unmarshal(registryContent, &registry)

	if err != nil {
		return Registry{}, err
	}

	return registry, nil
}

// Given a Registry and LockFile structs, return the list of dependencies from each LockFile remote
// that are present in the Registry but that have an incorrect URL.
//
// Example:
//
//	registry := Registry{
//	  Url: "https://packages.acme.io/",
//	  Dependencies: []string{
//	    "active_kafka",
//	    "cityhash",
//	  },
//	}
//
//	gemfile := GemfileLock{
//	  Remotes: map[string]Remote{
//	    "https://rubygems.org/": {
//	      Url: "https://rubygems.org/",
//	      Dependencies: map[string]Dependency{
//	        "active_kafka": { Name: "active_kafka" },
//	        "cityhash":     { Name: "cityhash" },
//	      },
//	    },
//	  },
//	}
//
//	registry.GetDependencyMismatches(gemfile)
//
//	=> ["active_kafka", "cityhash"]
//
// This happens since "active_kafka" and "cityhash" are dependencies present in the registry but found
// in a remote that contains a different URL from the registry itself.
func (registry *Registry) GetDependencyMismatches(lockFile interfaces.LockFile) map[string][]string {
	incorrectDependencies := make(map[string][]string, 0)

	for _, registryDependency := range registry.Dependencies {
		remotesUrls := lockFile.GetRemotesUrlsWithDependencyMismatch(registry.Url, registryDependency)

		for _, remoteUrl := range remotesUrls {
			incorrectDependencies[remoteUrl] = append(incorrectDependencies[remoteUrl], registryDependency)
		}
	}

	return incorrectDependencies
}
