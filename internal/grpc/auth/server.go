package auth

import (
	"context"
	"errors"
	sso_v1 "github.com/DenisRybkin/sso-protos/gen/go/sso"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"sso/internal/services/auth"
	"sso/internal/storage"
)

type Auth interface {
	Login(
		ctx context.Context,
		email string,
		password string,
		appID int,
	) (token string, err error)
	RegisterNewUser(
		ctx context.Context,
		email string,
		password string,
	) (userID int64, err error)
	IsAdmin(ctx context.Context, userID int64) (bool, error)
}

type serverAPI struct {
	sso_v1.UnimplementedAuthServer
	auth Auth
}

const (
	emptyValue = 0
)

func Register(gRPC *grpc.Server, auth Auth) {
	sso_v1.RegisterAuthServer(gRPC, &serverAPI{auth: auth})
}

func (s *serverAPI) Login(ctx context.Context, req *sso_v1.LoginRequest) (*sso_v1.LoginResponse, error) {
	if err := validateLogin(req); err != nil {
		return nil, err
	}
	token, err := s.auth.Login(ctx, req.GetEmail(), req.GetPassword(), int(req.GetAppId()))
	if err != nil {
		if errors.Is(err, auth.ErrInvalidCredentials) {
			return nil, status.Error(codes.InvalidArgument, "")
		}
		return nil, status.Errorf(codes.Internal, "internal error")
	}

	return &sso_v1.LoginResponse{Token: token}, nil
}

func (s *serverAPI) Register(ctx context.Context, req *sso_v1.RegisterRequest) (*sso_v1.RegisterResponse, error) {
	if err := validateRegister(req); err != nil {
		return nil, err
	}

	userID, err := s.auth.RegisterNewUser(ctx, req.GetEmail(), req.GetPassword())
	if err != nil {
		if errors.Is(err, auth.ErrUserExist) {
			return nil, status.Error(codes.AlreadyExists, "user already exist")
		}
		return nil, status.Error(codes.Internal, "internal error")
	}

	return &sso_v1.RegisterResponse{
		UserId: userID,
	}, nil
}

func (s *serverAPI) IsAdmin(ctx context.Context, req *sso_v1.IsAdminRequst) (*sso_v1.IsAdminResponse, error) {
	if err := validateIsAdmin(req); err != nil {
		return nil, err
	}

	isAdmin, err := s.auth.IsAdmin(ctx, req.GetUserId())
	if err != nil {
		if errors.Is(err, storage.ErrUserNotFound) {
			return nil, status.Error(codes.NotFound, "user not found")
		}
		return nil, status.Error(codes.Internal, "internal error")
	}
	return &sso_v1.IsAdminResponse{
		IsAdmin: isAdmin,
	}, nil

}

func validateLogin(req *sso_v1.LoginRequest) error {
	if req.GetEmail() == "" {
		return status.Error(codes.InvalidArgument, "empty email")
	}
	if req.GetPassword() == "" {
		return status.Error(codes.InvalidArgument, "empty password")
	}
	if req.GetAppId() == emptyValue {
		return status.Error(codes.InvalidArgument, "empty app id")
	}
	return nil
}

func validateRegister(req *sso_v1.RegisterRequest) error {
	if req.GetEmail() == "" {
		return status.Error(codes.InvalidArgument, "invalid email")
	}
	if req.GetPassword() == "" {
		return status.Error(codes.InvalidArgument, "invalid password")
	}
	return nil
}

func validateIsAdmin(req *sso_v1.IsAdminRequst) error {
	if req.GetUserId() == emptyValue {
		return status.Error(codes.InvalidArgument, "invalid user_id")
	}
	return nil
}
