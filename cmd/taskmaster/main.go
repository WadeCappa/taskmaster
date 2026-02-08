package main

import (
	"crypto/tls"
	"flag"
	"fmt"
	"log"
	"net"

	"github.com/WadeCappa/authmaster/pkg/go/authmaster/v1"
	"github.com/WadeCappa/taskmaster/internal/auth"
	"github.com/WadeCappa/taskmaster/internal/database"
	"github.com/WadeCappa/taskmaster/internal/server"
	"github.com/WadeCappa/taskmaster/pkg/go/tasks/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
)

var (
	port                 = flag.Int("port", 6100, "The server port")
	authHostname         = flag.String("auth-hostname", "localhost:50051", "The hostname for the auth server")
	authConnectionSecure = flag.Bool("auth-conn-secure", false, "Set this flag if the connection to the auth host is through TLS")
	psqlHostname         = flag.String("psql-hostname", "postgres://postgres:pass@postgres:5432/taskmaster_db", "Set this flag to the hostname of your postgres db")
)

func main() {
	flag.Parse()
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()

	log.Printf("looking at psql hostname of %s and authmaster hostname of %s", *psqlHostname, *authHostname)
	db := database.NewDatabase(*psqlHostname)

	if err := withConnection(*authHostname, *authConnectionSecure, func(ac authmaster.AuthmasterClient) error {
		auth := auth.NewAuth(ac)
		server_inst := server.NewServer(db, auth)
		taskspb.RegisterTasksServer(s, server_inst)

		log.Printf("server listening at %v", lis.Addr())
		if err := s.Serve(lis); err != nil {
			log.Fatalf("failed to serve: %v", err)
		}
		return nil
	}); err != nil {
		log.Fatalf("failed to connect to auth server: %v", err)
	}
}

func withConnection(hostname string, secure bool, consumer func(authmaster.AuthmasterClient) error) error {
	var creds credentials.TransportCredentials
	if secure {
		creds = credentials.NewTLS(&tls.Config{})
	} else {
		creds = insecure.NewCredentials()
	}
	conn, err := grpc.NewClient(hostname, grpc.WithTransportCredentials(creds))
	if err != nil {
		return fmt.Errorf("connecting to grpc server: %w", err)
	}
	defer conn.Close()
	if err := consumer(authmaster.NewAuthmasterClient(conn)); err != nil {
		return fmt.Errorf("consuming client: %w", err)
	}
	return nil
}
