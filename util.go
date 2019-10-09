package docspace

import (
	"os"
	"strings"
)

// GetGOPaths get all go paths
func GetGOPaths() []string {
	pathEnv := os.Getenv("GOPATH")
	paths := strings.Split(pathEnv, ":")
	for i := 0; i < len(paths); i++ {
		paths[i] = strings.TrimSpace(paths[i])
	}
	return paths
}
