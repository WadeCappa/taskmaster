package main

import (
	"context"
	"crypto/tls"
	"os"

	authmasterpb "github.com/WadeCappa/authmaster/pkg/go/authmaster/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"gopkg.in/yaml.v3"
)

type LoginCmd struct {
	Hostname     string `help:"Taskmaster server hostname." default:"localhost:6100"`
	AuthHostname string `help:"Authmaster server hostname." default:"localhost:50051"`
	Username     string `help:"Username." required:""`
	Password     string `help:"Password." required:""`
	Secure       bool   `help:"Use TLS for taskmaster." default:"true" negatable:""`
	AuthSecure   bool   `help:"Use TLS for authmaster." default:"true" negatable:""`
}

func (c *LoginCmd) Run(_ *Credentials) error {
	var tc credentials.TransportCredentials
	if c.AuthSecure {
		tc = credentials.NewTLS(&tls.Config{})
	} else {
		tc = insecure.NewCredentials()
	}
	conn, err := grpc.NewClient(c.AuthHostname, grpc.WithTransportCredentials(tc))
	if err != nil {
		return err
	}
	defer conn.Close()

	resp, err := authmasterpb.NewAuthmasterClient(conn).Login(
		context.Background(),
		&authmasterpb.LoginRequest{Username: c.Username, Password: c.Password},
	)
	if err != nil {
		return err
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	dir := home + "/.taskmaster"
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}

	data, err := yaml.Marshal(Credentials{
		Hostname: c.Hostname,
		Token:    resp.Token,
		Secure:   c.Secure,
	})
	if err != nil {
		return err
	}
	return os.WriteFile(dir+"/credentials.yml", data, 0600)
}
