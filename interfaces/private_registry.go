package interfaces

type PrivateRegistry interface {
	BuildRegistry(path string) error
	GetDependencyMismatches(lockFile LockFile) map[string][]string
}
