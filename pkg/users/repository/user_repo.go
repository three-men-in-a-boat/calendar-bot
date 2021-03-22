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

type OAuthError error

var (
	StateDoesNotExist            OAuthError = errors.New("state does not exist in redis")
	OAuthAccessTokenDoesNotExist OAuthError = errors.New("OAuth access token does not exist in redis")
	UserUnauthorized             OAuthError = errors.New("user exist in DB, but not authorized")
)

type UserEntityError error

var (
	UserDoesNotExist UserEntityError = errors.New("user does not exist")
)

const (
	TelegramOAuthPrefix = "telegram_id_"
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

func (us *UserRepository) SetTelegramUserIDByState(state string, telegramID int64, expire time.Duration) error {
	res := us.redisDB.Set(context.TODO(), state, fmt.Sprintf("%d", telegramID), expire)
	return res.Err()
}

func (us *UserRepository) GetTelegramUserIDByState(state string) (int64, error) {
	res := us.redisDB.Get(context.TODO(), state)

	if res.Err() == redis.Nil {
		return 0, StateDoesNotExist
	}

	return res.Int64()
}

func (us *UserRepository) SetOAuthAccessTokenByTelegramUserID(telegramID int64, oauthToken string, expire time.Duration) error {
	key := createOAuthRedisKeyForTelegram(telegramID)

	res := us.redisDB.Set(context.TODO(), key, oauthToken, expire)

	return res.Err()
}

func (us *UserRepository) GetOAuthAccessTokenByTelegramUserID(telegramID int64) (string, error) {
	key := createOAuthRedisKeyForTelegram(telegramID)
	res := us.redisDB.Get(context.TODO(), key)

	if res.Err() == redis.Nil {
		return "", OAuthAccessTokenDoesNotExist
	}

	return res.Result()
}

// Returns OAuthAccessToken
// Error types = error, OAuthError, UserEntityError
func (us *UserRepository) GetOAuthRefreshTokenByTelegramUserID(telegramID int64) (string, error) {
	var refreshToken sql.NullString
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
	case !refreshToken.Valid:
		return "", UserUnauthorized
	}

	return refreshToken.String, nil
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

func createOAuthRedisKeyForTelegram(telegramID int64) string {
	return TelegramOAuthPrefix + fmt.Sprintf("%d", telegramID)
}
