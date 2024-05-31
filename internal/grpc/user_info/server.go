package user_info

import (
	"context"
	"errors"
	ssov1 "github.com/bxiit/protos/gen/go/sso"
	"github.com/jinzhu/copier"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"sso/internal/data/models"
	"sso/internal/services/auth"
)

type UserInfo interface {
	GetUserInfo(ctx context.Context, token string) (*models.User, error)
}

type serverAPI struct {
	ssov1.UnimplementedUserInfoServer
	userinfo UserInfo
}

func Register(grpcServer *grpc.Server, ui UserInfo) {
	ssov1.RegisterUserInfoServer(grpcServer, &serverAPI{userinfo: ui})
}

func (s *serverAPI) GetUserInfo(ctx context.Context, in *ssov1.GetUserInfoRequest) (*ssov1.GetUserInfoResponse, error) {
	userInfo, err := s.userinfo.GetUserInfo(ctx, in.GetToken())
	if err != nil {
		switch {
		case errors.Is(err, auth.ErrNotValidJwt):
			return nil, status.Error(codes.PermissionDenied, "unknown user")
		}
		return nil, err
	}

	var uiResponse ssov1.User
	err = copier.Copy(&uiResponse, userInfo)
	if err != nil {
		return nil, err
	}
	return &ssov1.GetUserInfoResponse{User: &uiResponse}, nil
}
