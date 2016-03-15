package missinggo

import (
	"math/rand"
	"time"
)

func JitterDuration(average, plusMinus time.Duration) (ret time.Duration) {
	ret = average - plusMinus
	ret += time.Duration(rand.Int63n(2*int64(plusMinus) + 1))
	return
}
