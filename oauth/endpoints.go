package oauth

type Endpoint struct {
	AuthURL    string
	TokenURL   string
	ProfileURL string
}

var (
	FacebookEndpoint = Endpoint{
		AuthURL:    "https://www.facebook.com/dialog/oauth",
		TokenURL:   "https://graph.facebook.com/v2.3/oauth/access_token",
		ProfileURL: "https://graph.facebook.com/me",
	}
	GoogleEndpoint = Endpoint{
		AuthURL:    "https://accounts.google.com/o/oauth2/auth",
		TokenURL:   "https://accounts.google.com/o/oauth2/token",
		ProfileURL: "https://www.googleapis.com/oauth2/v2/userinfo",
	}
	PatreonEndpoint = Endpoint{
		AuthURL:    "https://www.patreon.com/oauth2/authorize",
		TokenURL:   "https://api.patreon.com/oauth2/token",
		ProfileURL: "https://api.patreon.com/oauth2/api/current_user",
	}
)
