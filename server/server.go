package server

import (
	"context"
	"net/http"
	"sync"

	"github.com/gofrs/uuid"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	usersv1 "github.com/johanbrandhorst/grpc-gateway-boilerplate/proto/users/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// Backend implements the protobuf interface
type Backend struct {
	mu    *sync.RWMutex
	users []*usersv1.User
}

// New initializes a new Backend struct.
func New() *Backend {
	return &Backend{
		mu: &sync.RWMutex{},
	}
}

// AddUser adds a user to the in-memory store.
func (b *Backend) AddUser(ctx context.Context, req *usersv1.AddUserRequest) (*usersv1.AddUserResponse, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	user := &usersv1.User{
		Id:         uuid.Must(uuid.NewV4()).String(),
		Email:      req.GetEmail(),
		CreateTime: timestamppb.Now(),
		Metadata:   req.GetMetadata(),
	}
	b.users = append(b.users, user)

	return &usersv1.AddUserResponse{
		User: user,
	}, nil
}

// GetUser gets a user from the store.
func (b *Backend) GetUser(ctx context.Context, req *usersv1.GetUserRequest) (*usersv1.GetUserResponse, error) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	for _, user := range b.users {
		if user.Id == req.GetId() {
			return &usersv1.GetUserResponse{User: user}, nil
		}
	}

	return nil, status.Errorf(codes.NotFound, "user with ID %q could not be found", req.GetId())
}

// ListUsers lists all users in the store.
func (b *Backend) ListUsers(_ *usersv1.ListUsersRequest, srv usersv1.UserService_ListUsersServer) error {
	b.mu.RLock()
	defer b.mu.RUnlock()

	for _, user := range b.users {
		err := srv.Send(&usersv1.ListUsersResponse{User: user})
		if err != nil {
			return err
		}
	}

	return nil
}

// UpdateUser updates properties on a user.
func (b *Backend) UpdateUser(ctx context.Context, req *usersv1.UpdateUserRequest) (*usersv1.UpdateUserResponse, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	var u *usersv1.User
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

	return &usersv1.UpdateUserResponse{User: u}, nil
}

// CustomErrorHandler defines the way we want errors to be shown to the users.
func CustomErrorHandler(ctx context.Context, mux *runtime.ServeMux, marshaler runtime.Marshaler, w http.ResponseWriter, req *http.Request, err error) {
	st := status.Convert(err)

	httpStatus := runtime.HTTPStatusFromCode(st.Code())
	w.WriteHeader(httpStatus)

	w.Write([]byte(st.Message()))
}
