package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"os"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"gopkg.in/yaml.v3"
)

const (
	credentialsFilePath = "/.taskmaster/credentials.yml"
)

type Credentials struct {
	Hostname string `yaml:"hostname"`
	Token    string `yaml:"token"`
	Secure   bool   `yaml:"secure"`
}

func loadCredentials() (*Credentials, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("getting home dir: %w", err)
	}
	data, err := os.ReadFile(home + credentialsFilePath)
	if err != nil {
		return nil, fmt.Errorf("reading credentials: %w", err)
	}
	var creds Credentials
	if err := yaml.Unmarshal(data, &creds); err != nil {
		return nil, fmt.Errorf("parsing credentials: %w", err)
	}
	return &creds, nil
}

func connect(creds *Credentials) (*grpc.ClientConn, error) {
	var tc credentials.TransportCredentials
	if creds.Secure {
		tc = credentials.NewTLS(&tls.Config{})
	} else {
		tc = insecure.NewCredentials()
	}
	return grpc.NewClient(creds.Hostname, grpc.WithTransportCredentials(tc))
}

func authCtx(creds *Credentials) context.Context {
	return metadata.NewOutgoingContext(
		context.Background(),
		metadata.Pairs("Authorization", creds.Token),
	)
}
