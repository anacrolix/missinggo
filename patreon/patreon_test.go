package patreon

import (
	"os"
	"testing"

	"github.com/go-quicktest/qt"
)

func TestParsePledges(t *testing.T) {
	f, err := os.Open("testdata/pledges")
	qt.Assert(t, qt.IsNil(err))
	defer f.Close()
	ps, err := ParsePledgesApiResponse(f)
	qt.Assert(t, qt.IsNil(err))
	qt.Check(t, qt.DeepEquals(ps, []Pledge{{
		Email:         "yonhyaro@gmail.com",
		EmailVerified: true,
		AmountCents:   200,
	}}))
}
