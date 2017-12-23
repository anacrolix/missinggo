package patreon

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type PledgesApiResponse struct {
	Pledges []struct {
		Attributes struct {
			AmountCents int `json:"amount_cents"`
		} `json:"attributes"`
		Relationships struct {
			Patron struct {
				Data struct {
					Id Id `json:"id"`
				} `json:"data"`
			} `json:"patron"`
		} `json:"relationships"`
	} `json:"data"`
	Included []ApiUser `json:"included"`
}

type ApiUser struct {
	Attributes struct {
		Email           string `json:"email"`
		IsEmailVerified bool   `json:"is_email_verified"`
	} `json:"attributes"`
	Id Id `json:"id"`
}

type Pledge struct {
	Email         string
	EmailVerified bool
	AmountCents   int
}

type Id string

func makeUserMap(par *PledgesApiResponse) (ret map[Id]*ApiUser) {
	ret = make(map[Id]*ApiUser, len(par.Included))
	for i := range par.Included {
		au := &par.Included[i]
		ret[au.Id] = au
	}
	return
}

func ParsePledgesApiResponse(r io.Reader) (ps []Pledge, err error) {
	var ar PledgesApiResponse
	err = json.NewDecoder(r).Decode(&ar)
	if err != nil {
		return
	}
	userMap := makeUserMap(&ar)
	for _, p := range ar.Pledges {
		u := userMap[p.Relationships.Patron.Data.Id]
		ps = append(ps, Pledge{
			Email:         u.Attributes.Email,
			EmailVerified: u.Attributes.IsEmailVerified,
			AmountCents:   p.Attributes.AmountCents,
		})
	}
	return
}

func GetCampaignPledges(campaign Id, userAccessToken string) (ret []Pledge, err error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("https://api.patreon.com/oauth2/api/campaigns/%s/pledges", campaign), nil)
	if err != nil {
		return
	}
	req.Header.Set("Authorization", "Bearer "+userAccessToken)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		err = fmt.Errorf("got http response code %d", resp.StatusCode)
		return
	}
	return ParsePledgesApiResponse(resp.Body)
}
