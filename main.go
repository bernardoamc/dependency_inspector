package main

import (
	"dependency_inspector/dependencies"
	"dependency_inspector/interfaces"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/exp/maps"
)

// Define a struct to hold our command line arguments for the 'analyze' command
type AnalyzeOptions struct {
	path     string
	registry string
	verbose  bool
	lang     string
}

// Define a struct to hold our command line arguments for the 'list' command
type RemoteOptions struct {
	path    string
	grep    string
	verbose bool
	lang    string
}

func usage() {
	helpMessage := `dependency_inspector: A tool for analyzing dependencies and listing remotes in .lock files.

Usage:

  dependency_inspector [command] [options]

Commands:
  analyze   Analyze .lock files for incorrect dependencies
  remotes   List every remote found in .lock files

Analyze command options:
  --path <path>       Specify the path to a directory containing .lock files or a single .lock file (REQUIRED)
  --registry <path>   Specify a registry file in JSON format (OPTIONAL, default: registry.json)
  --ruby              Analyze or list Ruby Gemfile.lock files (One of --ruby or --js is required)
  --js                Analyze or list yarn.lock files (One of --ruby or --js is required)
  --verbose           Print the dependencies parsed from each .lock file (OPTIONAL)

Remotes command options:
  --path <path>       Specify the path to a directory containing .lock files or a single .lock file (REQUIRED)
  --grep <string>     Specify a substring to match against remote URLs (OPTIONAL)
  --ruby              Analyze or list Ruby Gemfile.lock files (One of --ruby or --js is required)
  --js                Analyze or list yarn.lock files (One of --ruby or --js is required)
  --verbose           Print the dependencies parsed from each .lock file (OPTIONAL)

Examples:
  Analyze Gemfile.lock files:
    ./dependency_inspector analyze --path lock_files/ruby --registry registries/ruby.json --ruby

  Analyze yarn.lock files:
    ./dependency_inspector analyze --path lock_files/js --registry registries/js.json --js

  List remotes in Gemfile.lock files:
    ./dependency_inspector remotes --path lock_files/ruby --ruby

  List remotes in yarn.lock files:
    ./dependency_inspector remotes --path lock_files/js --js
`

	fmt.Fprint(os.Stderr, helpMessage)
}

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(1)
	}

	switch os.Args[1] {
	case "remotes":
		options := parseRemotesCommand(os.Args[2:])
		outputFolder := filepath.Join(filepath.Dir(os.Args[0]), "remotes_output")
		createDirectory(outputFolder)

		filenames, err := fetchLockFiles(options.path)
		if err != nil {
			fmt.Println("Error fetching lockfiles:", err)
			os.Exit(1)
		}

		listRemotes(filenames, outputFolder, options)
	case "analyze":
		options := parseAnalyzeCommand(os.Args[2:])
		outputFolder := filepath.Join(filepath.Dir(os.Args[0]), "analyze_output")
		createDirectory(outputFolder)

		filenames, err := fetchLockFiles(options.path)
		if err != nil {
			fmt.Println("Error fetching lockfiles:", err)
			os.Exit(1)
		}

		analyzeDependencies(filenames, outputFolder, options)
	case "--help":
		usage()
		os.Exit(0)
	default:
		fmt.Fprintf(os.Stderr, "Unknown command \"%s\", see --help for usage information\n", os.Args[1])
		os.Exit(1)
	}
}

func createDirectory(path string) {
	err := os.MkdirAll(path, 0755)
	if err != nil {
		fmt.Println("Error creating directory:", err)
		os.Exit(1)
	}
}

func parseRemotesCommand(args []string) RemoteOptions {
	remoteCmd := flag.NewFlagSet("remotes", flag.ExitOnError)
	pathFlag := remoteCmd.String("path", "", "Specify a lock file or folder path containing lock files")
	remoteMatchFlag := remoteCmd.String("grep", "", "Specify a substring to match against remote URLs")
	verboseFlag := remoteCmd.Bool("verbose", false, "Enable verbose output")
	rubyFlag := remoteCmd.Bool("ruby", false, "Parse Gemfile.lock files")
	jsFlag := remoteCmd.Bool("js", false, "Parse yarn.lock files")

	remoteCmd.Parse(args)

	if *pathFlag == "" {
		fmt.Println("You must specify a file or a folder to read your lock files from.")
		remoteCmd.Usage()
		os.Exit(1)
	}

	lang, err := getLang(*rubyFlag, *jsFlag)
	if err != nil {
		fmt.Println("Error inferring language:", err)
		os.Exit(1)
	}

	return RemoteOptions{
		path:    *pathFlag,
		grep:    *remoteMatchFlag,
		verbose: *verboseFlag,
		lang:    lang,
	}
}

func parseAnalyzeCommand(args []string) AnalyzeOptions {
	analyzeCmd := flag.NewFlagSet("analyze", flag.ExitOnError)
	pathFlag := analyzeCmd.String("path", "", "Specify a lock file or folder path containing lock files")
	registryFlag := analyzeCmd.String("registry", "registry.json", "Specify a registry file in JSON")
	verboseFlag := analyzeCmd.Bool("verbose", false, "Enable verbose output")
	rubyFlag := analyzeCmd.Bool("ruby", false, "Parse Gemfile.lock files")
	jsFlag := analyzeCmd.Bool("js", false, "Parse yarn.lock files")

	analyzeCmd.Parse(args)

	if *pathFlag == "" {
		fmt.Println("You must specify a file or a folder to read your lock files from.")
		analyzeCmd.Usage()
		os.Exit(1)
	}

	lang, err := getLang(*rubyFlag, *jsFlag)
	if err != nil {
		fmt.Println("Error inferring language:", err)
		os.Exit(1)
	}

	return AnalyzeOptions{
		path:     *pathFlag,
		registry: *registryFlag,
		verbose:  *verboseFlag,
		lang:     lang,
	}
}

func getLang(rubyFlag, jsFlag bool) (string, error) {
	if rubyFlag && jsFlag {
		return "", errors.New("a single language should be specified")
	} else if rubyFlag {
		return "ruby", nil
	} else if jsFlag {
		return "js", nil
	}

	return "", errors.New("no language specified")
}

func buildLockFile(lang string) interfaces.LockFile {
	if lang == "ruby" {
		return dependencies.NewGemfileLock()
	} else {
		return dependencies.NewYarnLock()
	}
}

func fetchLockFiles(path string) ([]string, error) {
	var lockFiles []string

	fileInfo, err := os.Stat(path)
	if err != nil {
		return nil, err
	}

	if fileInfo.IsDir() {
		files, err := os.ReadDir(path)
		if err != nil {
			return nil, err
		}

		for _, file := range files {
			if !file.IsDir() && strings.HasSuffix(file.Name(), ".lock") {
				lockFiles = append(lockFiles, filepath.Join(path, file.Name()))
			}
		}
	} else {
		if strings.HasSuffix(fileInfo.Name(), ".lock") {
			lockFiles = append(lockFiles, path)
		} else {
			return nil, errors.New("provided file is not a .lock file")
		}
	}

	return lockFiles, nil
}

func listRemotes(filenames []string, outputFolder string, options RemoteOptions) {
	matchedRemoteUrls := make(map[string]bool)

	for _, filename := range filenames {
		fmt.Printf("Parsing file: %s\n", filename)

		file, err := os.Open(filename)
		if err != nil {
			fmt.Println("Error opening file, skipping:", err)
			continue
		}
		defer file.Close()

		lockFile := buildLockFile(options.lang)
		lockFile.ParseLockFile(file)

		if options.verbose {
			lockFile.PrintRemotes()
		}

		for _, url := range lockFile.MatchRemoteUrls(options.grep) {
			matchedRemoteUrls[url] = true
		}
	}

	jsonData, err := json.MarshalIndent(maps.Keys(matchedRemoteUrls), "", "  ")
	if err != nil {
		fmt.Println("Error marshaling JSON:", err)
		return
	}

	outputPath := filepath.Join(outputFolder, "remotes.json")
	err = os.WriteFile(outputPath, jsonData, 0644)
	if err != nil {
		fmt.Println("Error writing file:", err)
		return
	}
}

func analyzeDependencies(filenames []string, outputFolder string, options AnalyzeOptions) {
	pkgRegistry, err := dependencies.BuildRegistry(options.registry)

	if err != nil {
		fmt.Println("Error building registry:", err)
		os.Exit(1)
	}

	fmt.Println(pkgRegistry)

	for _, filename := range filenames {
		// fmt.Printf("Parsing file: %s\n", filename)

		file, err := os.Open(filename)
		if err != nil {
			fmt.Println("Error opening file, skipping:", err)
			continue
		}
		defer file.Close()

		lockFile := buildLockFile(options.lang)
		lockFile.ParseLockFile(file)

		if options.verbose {
			lockFile.PrintRemotes()
		}

		dependencyMismatches := pkgRegistry.GetDependencyMismatches(lockFile)

		if len(dependencyMismatches) == 0 {
			continue
		}

		fmt.Printf("Dependency mismatches found for %s\n", filename)

		jsonData, err := json.MarshalIndent(dependencyMismatches, "", "  ")
		if err != nil {
			fmt.Println("Error marshaling JSON:", err)
			return
		}

		outputFile := fmt.Sprintf("%s_output.json", strings.TrimSuffix(filepath.Base(filename), ".lock"))
		outputPath := filepath.Join(outputFolder, outputFile)
		err = os.WriteFile(outputPath, jsonData, 0644)
		if err != nil {
			fmt.Println("Error writing file:", err)
			return
		}
	}
}
