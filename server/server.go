package server

import (
	"context"
	"sync"

	"github.com/gofrs/uuid"
	"github.com/golang/protobuf/ptypes/empty"

	pbExample "github.com/johanbrandhorst/grpc-gateway-boilerplate/proto"
)

type Backend struct {
	mu    *sync.RWMutex
	users []*pbExample.User
}

var _ pbExample.UserServiceServer = (*Backend)(nil)

func New() *Backend {
	return &Backend{
		mu: &sync.RWMutex{},
	}
}

func (b *Backend) AddUser(ctx context.Context, _ *empty.Empty) (*pbExample.User, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	user := &pbExample.User{
		Id: uuid.Must(uuid.NewV4()).String(),
	}
	b.users = append(b.users, user)

	return user, nil
}

func (b *Backend) ListUsers(_ *empty.Empty, srv pbExample.UserService_ListUsersServer) error {
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
