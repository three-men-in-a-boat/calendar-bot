package models

import "time"

type User struct {
	ID     int64
	UserID string

	MailUserEmail string

	MailAccessToken    string
	MailRefreshToken   string
	MailTokenExpiresIn time.Time

	TelegramUserId int64

	CreatedAt time.Time
}

//CREATE TABLE users
//(
//id                    BIGSERIAL PRIMARY KEY                              NOT NULL UNIQUE,
//mail_user_id          VARCHAR(128)                                       NOT NULL UNIQUE CHECK ( mail_user_id <> '' ),
//mail_user_email       VARCHAR(512)                                       NOT NULL UNIQUE CHECK ( mail_user_email <> '' ),
//
//mail_access_token     VARCHAR(128)                                       NOT NULL UNIQUE CHECK ( mail_access_token <> '' ),
//mail_refresh_token    VARCHAR(128)                                       NOT NULL UNIQUE CHECK ( mail_refresh_token <> '' ),
//mail_token_expires_in TIMESTAMP WITH TIME ZONE                           NOT NULL,
//
//telegram_user_id      BIGINT                                             NOT NULL UNIQUE,
//
//created_at            TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP NOT NULL
//--     nickname         VARCHAR(256)          NOT NULL,
//--     fullname         VARCHAR(512)          NOT NULL,
//);
