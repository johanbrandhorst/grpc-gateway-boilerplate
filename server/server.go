package server

import (
	"context"
	"sync"

	"github.com/gofrs/uuid"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/codes"

	pbExample "github.com/johanbrandhorst/grpc-gateway-boilerplate/proto"
)

// Backend implements the protobuf interface
type Backend struct {
	mu    *sync.RWMutex
	users []*pbExample.User
}

// New initializes a new Backend struct.
func New() *Backend {
	return &Backend{
		mu: &sync.RWMutex{},
	}
}

// AddUser adds a user to the in-memory store.
func (b *Backend) AddUser(ctx context.Context, req *pbExample.AddUserRequest) (*pbExample.User, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	user := &pbExample.User{
		Id: uuid.Must(uuid.NewV4()).String(),
		Email: req.GetEmail(),
	}
	b.users = append(b.users, user)

	return user, nil
}

// GetUser gets a user from the store.
func (b *Backend) GetUser(ctx context.Context, req *pbExample.GetUserRequest) (*pbExample.User, error) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	for _, user := range b.users {
		if user.Id == req.GetId() {
			return user, nil
		}
	}

	return nil, status.Errorf(codes.NotFound, "user with ID %q could not be found", req.GetId())
}

// ListUsers lists all users in the store.
func (b *Backend) ListUsers(_ *pbExample.ListUsersRequest, srv pbExample.UserService_ListUsersServer) error {
	b.mu.RLock()
	defer b.mu.RUnlock()

	for _, user := range b.users {
		err := srv.Send(user)
		if err != nil {
			return err
		}
	}

	return nil
}
