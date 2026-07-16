package rcon

import (
	"testing"

	"github.com/bobllor/assert"
	listenertest "github.com/bobllor/rcon-cli/listener/test"
)

func TestAuthenticate(t *testing.T) {
	li, err := listenertest.NewTcpListener()
	assert.Nil(t, err)
	defer li.Close()

	go li.HandleConnection()

	con, err := NewRcon(li.Addr().String())
	assert.Nil(t, err)
	defer con.Close()

	cases := []struct {
		Name     string
		Password string
		IsErr    bool
	}{
		{
			Name:     "Auth success",
			Password: listenertest.AuthPassword,
		},
		{
			Name:     "Auth fail",
			Password: "wrongpassword",
			IsErr:    true,
		},
	}

	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			err := con.Authenticate(c.Password)

			if c.IsErr {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
			}
		})
	}

}

func TestCommand(t *testing.T) {
	li, err := listenertest.NewTcpListener()
	assert.Nil(t, err)
	defer li.Close()

	go li.HandleConnection()

	con, err := NewRcon(li.Addr().String())
	assert.Nil(t, err)
	defer con.Close()

	res, err := con.Command("some command")
	assert.Nil(t, err)

	assert.True(t, len(res) != 0)
}
