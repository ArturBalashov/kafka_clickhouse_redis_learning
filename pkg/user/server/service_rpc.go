package server

import (
	"context"
	"errors"
	"github.com/go-redis/redis/v8"
	"github.com/jackc/pgx/v4/pgxpool"
	"go.uber.org/zap"
	"regexp"
	"time"

	"UserGrpcProj/models"
	pb "UserGrpcProj/pkg/user/service"
	"UserGrpcProj/repository/postgres"
)

type ServiceUser struct {
	pb.UserServer

	db     *pgxpool.Pool
	rdb    *redis.Client
	logger *zap.Logger
}

func NewUserService(db *pgxpool.Pool, logger *zap.Logger, rdb *redis.Client) *ServiceUser {
	return &ServiceUser{
		db:     db,
		rdb:    rdb,
		logger: logger,
	}
}
func (s *ServiceUser) AddUser(ctx context.Context, request *pb.AddUserRequest) (*pb.AddUserResponse, error) {
	login := request.GetLogin()

	temp := "^[a-zA-Z0-9]{5,}$"
	match, err := regexp.MatchString(temp, login)
	if err != nil {
		return nil, err
	}
	if !match {
		err := errors.New("wrong login")
		return nil, err
	}
	userRepository := postgres.NewUserRepo(ctx, s.db, s.logger)

	if userRepository.IsSetUser(login, 0) {
		return nil, errors.New("login already exists")
	}

	password := request.GetPassword()
	if len([]rune(password)) < 5 {
		err := errors.New("the password is too short")
		return nil, err
	}
	name := request.GetName()
	if len([]rune(name)) < 3 {
		err := errors.New("the name is too short")
		return nil, err
	}
	phone := request.GetPhone()
	if !IsPhoneValid(phone) {
		err := errors.New("wrong phone")
		return nil, err
	}
	user := models.UserInfo{
		Login:    login,
		Password: password,
		Name:     name,
		Phone:    phone,
	}

	ok, err := userRepository.AddUser(&user)
	status := &pb.AddUserResponse{
		Status: ok,
	}
	if err != nil {
		s.logger.Error("can't add user to repository", zap.Error(err))
		return status, err
	}

	s.logger.Info("new user was added", zap.String("login", user.Login))
	return status, nil
}
func (s *ServiceUser) RemoveUser(ctx context.Context, request *pb.RemoveUserRequest) (*pb.RemoveUserResponse, error) {
	removeId := request.GetId()
	if removeId <= 0 {
		return nil, errors.New("wrong request")
	}
	userRepository := postgres.NewUserRepo(ctx, s.db, s.logger)
	if userRepository.IsSetUser("", int(removeId)) == false {
		return nil, errors.New("id not found")
	}
	if removeId < 0 {
		return nil, errors.New("wrong id")
	}

	result, err := userRepository.RemoveUser(int(removeId))
	if err != nil {
		return nil, err
	}
	status := &pb.RemoveUserResponse{
		Status: result,
	}
	return status, nil
}

func (s *ServiceUser) UserList(ctx context.Context, request *pb.UserListRequest) (*pb.UserListResponse, error) {
	result := &pb.UserListResponse{}
	redisKey := "user_list_"

	filter := request.GetFilter()
	var phone, name, login string
	if filter != nil {
		phone = filter.Phone
		name = filter.Name
		login = filter.Login
		if phone != "" {
			redisKey += phone
		}
		if name != "" {
			redisKey += name
		}
		if login != "" {
			redisKey += login
		}
	}
	var resp *models.UserList
	if err := s.rdb.Get(ctx, redisKey).Scan(resp); err != nil && err.Error() != "redis: nil" {
		s.logger.Error("can't get user list from redis", zap.Error(err))
		return nil, err
	}
	if resp != nil {
		for _, u := range resp.User {
			result.UserList = append(result.UserList, &pb.UserInfo{
				Id:    int32(u.Id),
				Login: u.Password,
				Name:  u.Name,
				Phone: u.Phone,
			})
		}
		return result, nil
	}

	userRepository := postgres.NewUserRepo(ctx, s.db, s.logger)
	resp, err := userRepository.UserList(login, name, phone)
	if err != nil {
		return nil, err
	}

	if err := s.rdb.Set(ctx, redisKey, resp, 1*time.Minute).Err(); err != nil {
		s.logger.Error("can't set user list to redis", zap.Error(err))
		return nil, err
	}

	for _, respUser := range resp.User {
		result.UserList = append(result.UserList, &pb.UserInfo{
			Id:    int32(respUser.Id),
			Login: respUser.Login,
			Name:  respUser.Name,
			Phone: respUser.Phone,
		})
	}

	return result, nil
}

func IsPhoneValid(p string) bool {
	phoneRegex := regexp.MustCompile("^((8|\\+7)[\\- ]?)?(\\(?\\d{3}\\)?[\\- ]?)?[\\d\\- ]{7,10}$")
	if !phoneRegex.MatchString(p) || p == "" {
		return false
	}

	return true
}
