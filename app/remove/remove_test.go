package remove

import (
	"testing"

	"github.com/bobllor/assert"
	"github.com/bobllor/rcon-cli/app/test"
	"github.com/bobllor/rcon-cli/config"
)

func TestRemoveEntries(t *testing.T) {
	paths := test.NewAppPath(t)

	cfg, err := config.LoadConfigurationIfMissing(paths.Config)
	assert.Nil(t, err)

	entryNames := []string{"test1", "test2", "test3"}

	for _, name := range entryNames {
		cfg.AddEntry(name, config.RconEntry{})
	}

	assert.Nil(t, cfg.WriteFile(paths.Config))

	rc := NewRemoveCommand(paths)

	_, hasDeleted := rc.remove(cfg, entryNames)

	assert.True(t, hasDeleted)
	assert.True(t, len(cfg.RconEntries) == 0)
}
