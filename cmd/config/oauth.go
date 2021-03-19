package config

import "os"

const (
	EnvOAuthClientId               = "OAUTH_CLIENT_ID"
	EnvOAuthClientSecret           = "OAUTH_CLIENT_SECRET"
	EnvOAuthRedirectUri            = "OAUTH_REDIRECT_URI"
	EnvOAuthScope                  = "OAUTH_SCOPE"
	EnvOAuthTelegramBotRedirectUri = "OAUTH_TELEGRAM_BOT_REDIRECT_URI"
)

type OAuth struct {
	ClientID               string `valid:"hexadecimal"`
	ClientSecret           string `valid:"hexadecimal"`
	RedirectURI            string `valid:"requri"`
	Scope                  string `valid:"ascii"`
	TelegramBotRedirectURI string `valid:"requri"`
}

func LoadOAuthConfig() OAuth {
	clientID := os.Getenv(EnvOAuthClientId)
	clientSecret := os.Getenv(EnvOAuthClientSecret)
	redirectURI := os.Getenv(EnvOAuthRedirectUri)
	scope := os.Getenv(EnvOAuthScope)
	telegramBotRedirectUri := os.Getenv(EnvOAuthTelegramBotRedirectUri)

	// TODO(nickeskov): validate struct

	return OAuth{
		ClientID:               clientID,
		ClientSecret:           clientSecret,
		RedirectURI:            redirectURI,
		Scope:                  scope,
		TelegramBotRedirectURI: telegramBotRedirectUri,
	}
}

func (o *OAuth) ToEnv() map[string]string {
	return map[string]string{
		EnvOAuthClientId:               o.ClientID,
		EnvOAuthClientSecret:           o.ClientSecret,
		EnvOAuthRedirectUri:            o.RedirectURI,
		EnvOAuthScope:                  o.Scope,
		EnvOAuthTelegramBotRedirectUri: o.TelegramBotRedirectURI,
	}
}
