package repository

import (
	"database/sql"
	"fmt"
	"github.com/calendar-bot/pkg/customerrors"
	"github.com/calendar-bot/pkg/types"
	"github.com/pkg/errors"
	"strings"
)

type UserEntityError struct {
	error
}

var (
	UserDoesNotExist = UserEntityError{errors.New("user does not exist")}
)

type UserRepository struct {
	storage *sql.DB
}

func NewUserRepository(db *sql.DB) UserRepository {
	return UserRepository{
		storage: db,
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

func (us *UserRepository) GetUserEmailByTelegramUserID(telegramID int64) (string, error) {
	var email string
	err := us.storage.QueryRow(
		`SELECT mail_user_email FROM users WHERE telegram_user_id=$1`,
		telegramID,
	).Scan(
		&email,
	)

	switch {
	case err == sql.ErrNoRows:
		return "", UserDoesNotExist
	case err != nil:
		return "", errors.Wrapf(err, "failed to get user email by telegramID=%d", telegramID)
	}

	return email, nil
}

func (us *UserRepository) TryGetUsersEmailsByTelegramUserIDs(telegramIDs []int64) (emails []string, err error) {
	if len(telegramIDs) == 0 {
		return nil, nil
	}

	placeholders, err := postgresPlaceholdersForInSQLExpression(len(telegramIDs))
	if err != nil {
		return nil, errors.WithStack(err)
	}

	query := fmt.Sprintf(
		"SELECT mail_user_email FROM users WHERE telegram_user_id IN (%s)",
		placeholders,
	)
	queryArgs := make([]interface{}, 0, len(telegramIDs))
	for _, id := range telegramIDs {
		queryArgs = append(queryArgs, id)
	}

	rows, err := us.storage.Query(query, queryArgs...)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to perform SQL query in TryGetUsersEmailsByTelegramUserIDs")
	}
	defer func() {
		err = customerrors.HandleCloser(err, rows)
	}()

	for rows.Next() {
		var email string
		if err := rows.Scan(&email); err != nil {
			return nil, errors.Wrap(err, "error while scanning user emails")
		}
		emails = append(emails, email)
	}

	return emails, nil
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

func postgresPlaceholdersForInSQLExpression(placeholdersCount int) (string, error) {
	if placeholdersCount <= 0 {
		return "", errors.Errorf(
			"failed generate placeholders for IN SQL expression, placeholdersCount=%d",
			placeholdersCount,
		)
	}

	var builder strings.Builder

	builder.WriteString("$1")
	for i := 2; i <= placeholdersCount; i++ {
		if _, err := fmt.Fprintf(&builder, ",$%d", i); err != nil {
			// nickeskov: it's impossible case...
			return "", errors.Wrap(err, "failed to write into strings.Builder")
		}
	}

	return builder.String(), nil
}
