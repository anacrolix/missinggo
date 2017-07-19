package filecache

import "testing"

func TestLRU(t *testing.T) {
	testPolicy(t, &lru{})
}
