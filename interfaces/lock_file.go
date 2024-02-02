package interfaces

import "io"

type LockFile interface {
	ParseLockFile(file io.Reader)
	PrintRemotes()
	MatchRemoteUrls(grep string) []string
	GetRemotesUrlsWithDependencyMismatch(registryUrl, dependency string) []string
}
