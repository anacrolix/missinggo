package oauth

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/anacrolix/missinggo/patreon"
)

func SimpleParser(r *http.Response) (UserProfile, error) {
	var sup simpleUserProfile
	err := json.NewDecoder(r.Body).Decode(&sup)
	return sup, err
}

type Provider struct {
	Client   *Client
	Endpoint *Endpoint
}

type Wrapper struct {
	Scope         string
	Provider      Provider
	ProfileParser func(*http.Response) (UserProfile, error)
}

func (me Wrapper) GetAuthURL(redirectURI, state string) string {
	return me.Provider.GetAuthURL(redirectURI, state, me.Scope)
}

func (me Wrapper) FetchUser(accessToken string) (up UserProfile, err error) {
	resp, err := me.Provider.FetchUser(accessToken)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	return me.ProfileParser(resp)
}

type Client struct {
	ID     string
	Secret string
}

func (me *Provider) GetAuthURL(redirectURI, state, scope string) string {
	params := []string{
		"client_id", me.Client.ID,
		"response_type", "code",
		"redirect_uri", redirectURI,
		"state", state,
		// This will ask again for the given scopes if they're not provided.
		"auth_type", "rerequest",
	}
	if scope != "" {
		params = append(params, "scope", scope)
	}
	return renderEndpointURL(me.Endpoint.AuthURL, params...)
}

func (me *Provider) ExchangeCode(code string, redirectURI string) (accessToken string, err error) {
	v := url.Values{
		"client_id":     {me.Client.ID},
		"redirect_uri":  {redirectURI},
		"client_secret": {me.Client.Secret},
		"code":          {code},
		"grant_type":    {"authorization_code"},
	}
	resp, err := http.Post(me.Endpoint.TokenURL, "application/x-www-form-urlencoded", bytes.NewBufferString(v.Encode()))
	if err != nil {
		return
	}
	var buf bytes.Buffer
	io.Copy(&buf, resp.Body)
	resp.Body.Close()
	var msg map[string]interface{}
	err = json.NewDecoder(&buf).Decode(&msg)
	if err != nil {
		return
	}
	defer func() {
		r := recover()
		if r == nil {
			return
		}
		err = fmt.Errorf("bad access_token field in %q: %s", msg, r)
	}()
	accessToken = msg["access_token"].(string)
	return
}

type simpleUserProfile struct {
	Id         string `json:"id"`
	EmailField string `json:"email"`
}

var _ UserProfile = simpleUserProfile{}

func (me simpleUserProfile) IsEmailVerified() bool {
	return true
}

func (me simpleUserProfile) Email() string {
	return me.EmailField
}

type UserProfile interface {
	IsEmailVerified() bool
	Email() string
}

// TODO: Allow fields to be specified.
func (me *Provider) FetchUser(accessToken string) (*http.Response, error) {
	return http.Get(renderEndpointURL(
		me.Endpoint.ProfileURL,
		"fields", "email",
		"access_token", accessToken,
	))
}

type PatreonUserProfile struct {
	Data patreon.ApiUser `json:"data"`
}

var _ UserProfile = PatreonUserProfile{}

func (me PatreonUserProfile) IsEmailVerified() bool {
	return me.Data.Attributes.IsEmailVerified
}

func (me PatreonUserProfile) Email() string {
	return me.Data.Attributes.Email
}

func renderEndpointURL(endpoint string, params ...string) string {
	u, err := url.Parse(endpoint)
	if err != nil {
		panic(err)
	}
	v := make(url.Values, len(params)/2)
	for i := 0; i < len(params); i += 2 {
		v.Set(params[i], params[i+1])
	}
	u.RawQuery = v.Encode()
	return u.String()
}
