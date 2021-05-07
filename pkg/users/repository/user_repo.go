package repository

import (
	"database/sql"
	"github.com/calendar-bot/pkg/types"
	"github.com/go-redis/redis/v8"
	"github.com/pkg/errors"
)

type UserEntityError error

var (
	UserDoesNotExist UserEntityError = errors.New("user does not exist")
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

// GetOAuthRefreshTokenByTelegramUserID returns OAuthAccessToken
// Error types = error, UserEntityError
func (us *UserRepository) GetOAuthRefreshTokenByTelegramUserID(telegramID int64) (string, error) {
	var refreshToken string
	err := us.storage.QueryRow(
		`SELECT u.mail_refresh_token FROM users AS u WHERE u.telegram_user_id = $1`,
		telegramID,
	).Scan(
		&refreshToken,
	)

	switch {
	case err == sql.ErrNoRows:
		return "", UserDoesNotExist
	case err != nil:
		return "", errors.Wrapf(err, "failed to get mail refresh token by telegramID=%d", telegramID)
	}

	return refreshToken, nil
}

func (us *UserRepository) CreateUser(user types.TelegramDBUser) error {

	_, err := us.storage.Exec(`
			INSERT INTO users(
							  mail_user_id,
							  mail_user_email,
			                  mail_refresh_token,
							  telegram_user_id)
			VALUES ($1, $2, $3, $4)`,
		user.MailUserID,
		user.MailUserEmail,
		user.MailRefreshToken,
		user.TelegramUserId,
	)

	if err != nil {
		return errors.Wrapf(err, "cannot create user=%v", user)
	}

	return nil
}

func (us *UserRepository) DeleteUserByTelegramUserID(telegramID int64) error {

	err := us.storage.QueryRow(
		`DELETE FROM users WHERE telegram_user_id=$1 RETURNING telegram_user_id`,
		telegramID,
	).Scan(
		&telegramID,
	)

	switch {
	case err == sql.ErrNoRows:
		return UserDoesNotExist
	case err != nil:
		return errors.Wrapf(err, "cannot delete user with telegramUserID=%d", telegramID)
	}

	return nil
}
