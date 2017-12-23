package patreon

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParsePledges(t *testing.T) {
	f, err := os.Open("testdata/pledges")
	require.NoError(t, err)
	defer f.Close()
	ps, err := ParsePledgesApiResponse(f)
	require.NoError(t, err)
	assert.EqualValues(t, []Pledge{{
		Email:         "yonhyaro@gmail.com",
		EmailVerified: true,
		AmountCents:   200,
	}}, ps)
}
