package server

import (
	"context"
	"net/http"
	"sync"

	"github.com/gofrs/uuid"
	"github.com/golang/protobuf/ptypes"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

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
		Id:         uuid.Must(uuid.NewV4()).String(),
		Email:      req.GetEmail(),
		CreateTime: ptypes.TimestampNow(),
		Metadata:   req.GetMetadata(),
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

// UpdateUser updates properties on a user.
func (b *Backend) UpdateUser(ctx context.Context, req *pbExample.UpdateUserRequest) (*pbExample.User, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	var u *pbExample.User
	for _, user := range b.users {
		if user.Id == req.GetUser().GetId() {
			u = user
			break
		}
	}
	if u == nil {
		return nil, status.Errorf(codes.NotFound, "user with ID %q could not be found", req.GetUser().GetId())
	}

	for _, path := range req.GetUpdateMask().GetPaths() {
		switch path {
		case "email":
			u.Email = req.GetUser().GetEmail()
		default:
			return nil, status.Errorf(codes.InvalidArgument, "cannot update field %q on user", path)
		}
	}

	return u, nil
}

// CustomErrorHandler defines the way we want errors to be shown to the users.
func CustomErrorHandler(ctx context.Context, mux *runtime.ServeMux, marshaler runtime.Marshaler, w http.ResponseWriter, req *http.Request, err error) {
	st := status.Convert(err)

	httpStatus := runtime.HTTPStatusFromCode(st.Code())
	w.WriteHeader(httpStatus)

	w.Write([]byte(st.Message()))
}
