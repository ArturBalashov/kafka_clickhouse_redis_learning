package postgres

import (
	"context"
	"crypto/md5"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"go.uber.org/zap"

	"UserGrpcProj/models"
	"UserGrpcProj/repository"
)

type userRepoPostgres struct {
	db     *pgxpool.Pool
	ctx    context.Context
	logger *zap.Logger
}

func NewUserRepo(ctx context.Context, db *pgxpool.Pool, logger *zap.Logger) repository.UserRepository {
	return &userRepoPostgres{
		db:     db,
		ctx:    ctx,
		logger: logger,
	}
}
func (u *userRepoPostgres) AddUser(user *models.UserInfo) (bool, error) {
	if _, err := u.db.Exec(u.ctx, "INSERT INTO users (login, password, name, phone) VALUES ($1,$2,$3,$4)", user.Login, GetMD5Hash(user.Password), user.Name, user.Phone); err != nil {
		return false, err
	}
	return true, nil
}

func (u *userRepoPostgres) RemoveUser(id int) (bool, error) {
	if _, err := u.db.Exec(u.ctx, "DELETE FROM users WHERE id=$1", id); err != nil {
		u.logger.Error("can't delete user", zap.Error(err))
		return false, err
	}

	return true, nil
}

func (u *userRepoPostgres) UserList(login string, name string, phone string) (*models.UserList, error) {
	rows, err := u.db.Query(u.ctx, "SELECT id, login, name, phone FROM users WHERE login ILIKE $1 OR name ILIKE $2 OR phone ILIKE $3", fmt.Sprintf("%s", login), fmt.Sprintf("%s", name), fmt.Sprintln("%s", phone))
	if err != nil {
		return nil, err
	}

	respUserList := models.UserList{}
	for rows.Next() {
		var userIdRow sql.NullInt32
		var userLoginRow, userNameRow, userPhoneRow sql.NullString
		if err := rows.Scan(&userIdRow, &userLoginRow, &userNameRow, &userPhoneRow); err != nil {
			return nil, err
		}

		respUserList.User = append(respUserList.User, &models.UserInfo{
			Id:    int(userIdRow.Int32),
			Login: userLoginRow.String,
			Name:  userNameRow.String,
			Phone: userPhoneRow.String,
		})
	}

	return &respUserList, nil
}

func (u *userRepoPostgres) IsSetUser(byLogin string, byId int) bool {
	var id sql.NullInt64
	if byLogin != "" {
		if err := u.db.QueryRow(u.ctx, "SELECT id FROM users WHERE login=$1", byLogin).Scan(&id); err != nil {
			if !errors.Is(err, pgx.ErrNoRows) {
				u.logger.Error("error", zap.Error(err))
			}
			return false
		}
		return true
	}
	if byId > 0 {
		if err := u.db.QueryRow(u.ctx, "SELECT id FROM users WHERE id=$1", byId).Scan(&id); err != nil {
			if !errors.Is(err, pgx.ErrNoRows) {
				u.logger.Error("error", zap.Error(err))
			}
			return false
		}
	}
	return true
}
func GetMD5Hash(text string) string {
	hash := md5.Sum([]byte(text))
	return hex.EncodeToString(hash[:])
}
