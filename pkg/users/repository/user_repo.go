package repository

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/calendar-bot/pkg/types"
	"github.com/go-redis/redis/v8"
	"github.com/pkg/errors"
	"time"
)

var (
	StateDoesNotExist = errors.New("state does not exist in redis")
)

type UserRepository struct {
	storage *sql.DB
	redisDB *redis.Client
}

func NewUserRepository(db *sql.DB, client *redis.Client) UserRepository {
	return UserRepository{
		storage: db,
		redisDB: client,
	}
}

func (us *UserRepository) SetTelegramIDByState(state string, telegramID int64, expire time.Duration) error {
	res := us.redisDB.Set(context.TODO(), state, fmt.Sprintf("%d", telegramID), expire)
	return res.Err()
}

func (us *UserRepository) GetTelegramIDByState(state string) (int64, error) {
	res := us.redisDB.Get(context.TODO(), state)

	if res.Err() == redis.Nil {
		return 0, StateDoesNotExist
	}

	return res.Int64()
}

func (us *UserRepository) CreateUser(user types.User) error {

	_, err := us.storage.Exec(`
			INSERT INTO users(
							  mail_user_id,
							  mail_user_email, 
							  mail_access_token,
							  mail_refresh_token,
							  mail_token_expires_in,
							  telegram_user_id)
			VALUES ($1, $2, $3, $4, $5, $6)`,
		user.UserID,
		user.MailUserEmail,
		user.MailAccessToken,
		user.MailRefreshToken,
		user.MailTokenExpiresIn,
		user.TelegramUserId,
	)

	if err != nil {
		return errors.Wrapf(err, "cannot create user=%v", user)
	}

	return nil
}
