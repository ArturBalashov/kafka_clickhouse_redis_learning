package repository

import "UserGrpcProj/models"

type UserRepository interface {
	AddUser(user *models.UserInfo) (bool, error)
	RemoveUser(int) (bool, error)
	UserList(string, string, string) (*models.UserList, error)
	IsSetUser(string, int) bool
}
