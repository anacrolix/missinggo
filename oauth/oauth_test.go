package oauth

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDecodePatreonUserProfile(t *testing.T) {
	var pup PatreonUserProfile
	err := json.Unmarshal([]byte(
		`{
    "data": {
        "attributes": {
            "about": null,
            "created": "2017-05-12T12:49:31+00:00",
            "discord_id": null,
            "email": "anacrolix@gmail.com",
            "facebook": null,
            "facebook_id": "10155425587018447",
            "first_name": "Matt",
            "full_name": "Matt Joiner",
            "gender": 0,
            "has_password": false,
            "image_url": "https://c3.patreon.com/2/patreon-user/wS20eHsYaLMqJeDyL5wyK0egvcXDRNdT28JvjeREJ5T80te19Cmn1YZxZyzd2qab.jpeg?t=2145916800&w=400&v=506YL5JlU7aaQH-QyEaRXyWoXFs4ia-vcSjjZuv-dXY%3D",
            "is_deleted": false,
            "is_email_verified": true,
            "is_nuked": false,
            "is_suspended": false,
            "last_name": "Joiner",
            "social_connections": {
                "deviantart": null,
                "discord": null,
                "facebook": null,
                "spotify": null,
                "twitch": null,
                "twitter": null,
                "youtube": null
            },
            "thumb_url": "https://c3.patreon.com/2/patreon-user/wS20eHsYaLMqJeDyL5wyK0egvcXDRNdT28JvjeREJ5T80te19Cmn1YZxZyzd2qab.jpeg?h=100&t=2145916800&w=100&v=SI72bzI4XB5mX0dyfqeZ-Nn4BNTz9FYRSgZ8pLipARg%3D",
            "twitch": null,
            "twitter": null,
            "url": "https://www.patreon.com/anacrolix",
            "vanity": "anacrolix",
            "youtube": null
        },
        "id": "6126463",
        "relationships": {
            "pledges": {
                "data": []
            }
        },
        "type": "user"
    },
    "links": {
        "self": "https://api.patreon.com/user/6126463"
    }
}`), &pup)
	require.NoError(t, err)
	assert.EqualValues(t, "anacrolix@gmail.com", pup.Data.Attributes.Email)
	assert.True(t, pup.Data.Attributes.IsEmailVerified)
}
