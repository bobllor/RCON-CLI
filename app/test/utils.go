package test

import (
	"path/filepath"
	"testing"

	"github.com/bobllor/rcon-cli/app/utils/paths"
)

// NewAppPath returns an AppPath set to the test temp directory.
//
// The caller is responsible for the creation of the paths.
//
// The Home field will be root.
func NewAppPath(t *testing.T) paths.AppPath {
	dir := t.TempDir()

	appPaths := paths.AppPath{
		Home: dir,
	}

	appPaths.Config = filepath.Join(dir, paths.ConfigPathRel)
	appPaths.Runtime = filepath.Join(dir, paths.RuntimePathRel)
	appPaths.Log = filepath.Join(dir, paths.LogPathRel)

	return appPaths
}
