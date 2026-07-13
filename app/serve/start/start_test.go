package ipcstart

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/bobllor/assert"
	"github.com/bobllor/rcon/app/test"
	"github.com/bobllor/rcon/app/utils/files"
)

func TestCheckAndRemoveStaleService(t *testing.T) {
	paths := test.NewAppPath(t)

	errs := paths.MkdirAll()
	assert.Nil(t, errs)

	addr := filepath.Join(paths.Runtime, files.SocketFile)

	cases := []struct {
		name    string
		addr    string
		pidFile string
		makePid bool
	}{
		{
			name:    "Removal (No PID File)",
			addr:    addr,
			pidFile: "doesnotexist",
		},
		{
			name:    "Removal (PID file)",
			addr:    addr,
			pidFile: filepath.Join(paths.Runtime, files.PidFile),
			makePid: true,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			err := os.WriteFile(c.addr, []byte{}, 0o0744)
			assert.Nil(t, err)

			if c.makePid {
				err := os.WriteFile(c.pidFile, []byte{}, 0o0744)
				assert.Nil(t, err)
			}

			startCmd := NewIpcStartCommand(addr, c.pidFile, paths)
			startCmd.checkAndRemoveStaleService(c.pidFile, c.addr)

			_, err = os.Stat(c.addr)
			assert.NotNil(t, err)

			if c.makePid {
				_, err := os.Stat(c.pidFile)
				assert.NotNil(t, err)
			}
		})
	}
}
