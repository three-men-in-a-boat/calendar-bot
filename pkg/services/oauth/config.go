package oauth

import (
	"fmt"
	"github.com/pkg/errors"
	"os"
	"time"
)

const (
	EnvOAuthClientId               = "OAUTH_CLIENT_ID"
	EnvOAuthClientSecret           = "OAUTH_CLIENT_SECRET"
	EnvOAuthRedirectUri            = "OAUTH_REDIRECT_URI"
	EnvOAuthScope                  = "OAUTH_SCOPE"
	EnvOAuthTelegramBotRedirectUri = "OAUTH_TELEGRAM_BOT_REDIRECT_URI"
	EnvOAuthLinkExpireIn           = "OAUTH_LINK_EXPIRE_IN"
)

const oauthLinkExpireInDefault = 15 * time.Minute

type Config struct {
	ClientID               string        `valid:"hexadecimal"`
	ClientSecret           string        `valid:"hexadecimal"`
	RedirectURI            string        `valid:"requri"`
	Scope                  string        `valid:"ascii"`
	TelegramBotRedirectURI string        `valid:"requri"`
	LinkExpireIn           time.Duration `valid:"-"`
}

func LoadOAuthConfig() (Config, error) {
	clientID := os.Getenv(EnvOAuthClientId)
	clientSecret := os.Getenv(EnvOAuthClientSecret)
	redirectURI := os.Getenv(EnvOAuthRedirectUri)
	scope := os.Getenv(EnvOAuthScope)
	telegramBotRedirectUri := os.Getenv(EnvOAuthTelegramBotRedirectUri)

	linkExpireIn := oauthLinkExpireInDefault
	if linkExpireInStr := os.Getenv(EnvOAuthLinkExpireIn); linkExpireInStr != "" {
		expire, err := time.ParseDuration(linkExpireInStr)
		if err != nil {
			return Config{}, errors.WithMessagef(
				err,
				"failed to parse %s environment variable as time.Duration",
				EnvOAuthLinkExpireIn,
			)
		}
		if expire <= 0 {
			return Config{}, errors.Errorf("%s duration must be greater than zero", EnvOAuthLinkExpireIn)
		}
		linkExpireIn = expire
	}

	// TODO(nickeskov): validate struct

	return Config{
		ClientID:               clientID,
		ClientSecret:           clientSecret,
		RedirectURI:            redirectURI,
		Scope:                  scope,
		TelegramBotRedirectURI: telegramBotRedirectUri,
		LinkExpireIn:           linkExpireIn,
	}, nil
}

func (o *Config) ToEnv() map[string]string {
	return map[string]string{
		EnvOAuthClientId:               o.ClientID,
		EnvOAuthClientSecret:           o.ClientSecret,
		EnvOAuthRedirectUri:            o.RedirectURI,
		EnvOAuthScope:                  o.Scope,
		EnvOAuthTelegramBotRedirectUri: o.TelegramBotRedirectURI,
		EnvOAuthLinkExpireIn:           fmt.Sprintf("%s", o.LinkExpireIn),
	}
}
