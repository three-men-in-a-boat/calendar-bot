package oauth

import (
	"fmt"
	"github.com/pkg/errors"
	"net/url"
)

const (
	paramAccessToken  = "access_token"
	paramRefreshToken = "refresh_token"
	paramClientID     = "client_id"
	paramClientSecret = "client_secret"
	paramResponseType = "response_type"
	paramRedirectURI  = "redirect_uri"
	paramGrantType    = "grant_type"
	paramScope        = "scope"
	paramState        = "state"
	paramCode         = "code"
)

type ObtainTokensResponse struct {
	ExpiresInSeconds int64  `json:"expires_in"`
	AccessToken      string `json:"access_token"`
	RefreshToken     string `json:"refresh_token"`
	TokenType        string `json:"token_type"`
	*APIResponseErr
}

type RenewAccessTokenResponse struct {
	ExpiresInSeconds int64  `json:"expires_in"`
	AccessToken      string `json:"access_token"`
	*APIResponseErr
}

// Gender: m - male, f - female

type UserInfoResponse struct {
	ID        string `json:"id"`
	Gender    string `json:"gender"`
	Name      string `json:"name"`
	Nickname  string `json:"nickname"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Locale    string `json:"locale"`
	Email     string `json:"email"`
	Birthday  string `json:"birthday"`
	Image     string `json:"image"`
	*APIResponseErr
	//ClientID  string
}

type link struct {
	linkUrl     *url.URL
	queryParams url.Values
}

func (l *link) String() string {
	urlValues := make(url.Values, len(l.queryParams))
	for key, values := range l.queryParams {
		for i := range values {
			urlValues.Add(key, values[i])
		}
	}
	return fmt.Sprintf("%s?%s", l.linkUrl.String(), urlValues.Encode())
}

type LoginLink struct {
	link
}

func newLoginLink(conf *Config, stateKey string) (LoginLink, error) {
	rawUrl := fmt.Sprintf("%s/login", conf.HostWithScheme)
	loginUrl, err := url.ParseRequestURI(rawUrl)
	if err != nil {
		return LoginLink{}, errors.Wrapf(err, "cannot parse loginUrl from string '%s'", rawUrl)
	}

	return LoginLink{
		link{
			linkUrl: loginUrl,
			queryParams: url.Values{
				paramClientID:     []string{conf.ClientID},
				paramResponseType: []string{"code"},
				paramScope:        []string{conf.Scope},
				paramRedirectURI:  []string{conf.RedirectURI},
				paramState:        []string{stateKey},
			},
		},
	}, nil
}

func (l *LoginLink) State() string {
	return l.queryParams[paramState][0]
}

func newUserInfoLink(conf *Config, accessToken string) (link, error) {
	rawUrl := fmt.Sprintf("%s/userinfo", conf.HostWithScheme)
	userInfoUrl, err := url.ParseRequestURI(rawUrl)
	if err != nil {
		return link{},
			errors.Wrapf(err, "cannot parse userInfoUrl from string '%s'", rawUrl)
	}

	return link{
		linkUrl: userInfoUrl,
		queryParams: url.Values{
			paramAccessToken: []string{accessToken},
		},
	}, nil
}

type tokensPostForm struct {
	tokensUrl *url.URL
	values    url.Values
}

func newObtainTokensPostForm(conf *Config, authCode string) (tokensPostForm, error) {
	rawUrl := fmt.Sprintf("%s/token", conf.HostWithScheme)
	obtainTokensUrl, err := url.ParseRequestURI(rawUrl)
	if err != nil {
		return tokensPostForm{},
			errors.Wrapf(err, "cannot parse obtain tokens url from string '%s'", rawUrl)
	}

	return tokensPostForm{
		tokensUrl: obtainTokensUrl,
		values: url.Values{
			paramCode:         []string{authCode},
			paramGrantType:    []string{"authorization_code"},
			paramRedirectURI:  []string{conf.RedirectURI},
			paramClientID:     []string{conf.ClientID},
			paramClientSecret: []string{conf.ClientSecret},
		},
	}, nil
}

func newRenewAccessTokenPostForm(conf *Config, refreshToken string) (tokensPostForm, error) {
	rawUrl := fmt.Sprintf("%s/token", conf.HostWithScheme)
	renewAccessTokenUrl, err := url.ParseRequestURI(rawUrl)
	if err != nil {
		return tokensPostForm{},
			errors.Wrapf(err, "cannot parse renewAccessTokenUrl from string '%s'", rawUrl)
	}

	return tokensPostForm{
		tokensUrl: renewAccessTokenUrl,
		values: url.Values{
			paramClientID:     []string{conf.ClientID},
			paramGrantType:    []string{"refresh_token"},
			paramRefreshToken: []string{refreshToken},
		},
	}, nil
}
