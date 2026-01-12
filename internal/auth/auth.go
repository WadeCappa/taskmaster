package auth

import (
	"context"
	"errors"
	"fmt"

	"github.com/WadeCappa/authmaster/pkg/go/authmaster/v1"
	"google.golang.org/grpc/metadata"
)

type Auth struct {
	client authmaster.AuthmasterClient
}

func NewAuth(client authmaster.AuthmasterClient) *Auth {
	return &Auth{
		client: client,
	}
}

func (a *Auth) GetUserId(ctx context.Context) (UserId, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return 0, errors.New("no authorization header for request")
	}

	newContext := metadata.NewOutgoingContext(ctx, md)
	resp, err := a.client.TestAuth(newContext, &authmaster.TestAuthRequest{})
	if err != nil {
		return 0, fmt.Errorf("testing auth-header: %w", err)
	}

	return UserId(resp.GetUserId()), nil
}
