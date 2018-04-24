package proj5

import (
	"testing"
)

// Unit tests for our cache
func TestCache(t *testing.T) {
	handle := CacheHandle{
		make(chan CacheReq),
		make(chan CacheResp),
	}
	go Cache(handle)

	var reqID int64 = 0
	// Write the value 42 with key 1 to the cache
	handle.ReqQ <- CacheReq{true, 1, 42, reqID}
	reqID++
	// It better be in there when we ask for it!
	if resp := read(1, handle, t, &reqID); !resp.Exists || resp.Val != 42 {
		t.Errorf("Expected (true, %d), got %v", 42, resp)
	}
	// Maybe the cache accidentally deleted it? Reads shouldn't be destructive.
	if resp := read(1, handle, t, &reqID); !resp.Exists || resp.Val != 42 {
		t.Errorf("Expected (true, %d) on second try, got %v", 42, resp)
	}

	// We've never written a "2" to the cache, we shouldn't get anything back
	if resp := read(2, handle, t, &reqID); resp.Exists {
		t.Errorf("Requesting non-existent key: Expected (false, 0), got %v", resp)
	}
	// Just asking for a key shouldn't change the cache
	if resp := read(2, handle, t, &reqID); resp.Exists {
		t.Errorf("Requesting non-existent key (again): Expected (false, 0), got %v", resp)
	}

	// Overwrite the value at key 1 (it's now 17)
	handle.ReqQ <- CacheReq{true, 1, 17, reqID}
	reqID++
	if resp := read(1, handle, t, &reqID); !resp.Exists || resp.Val != 17 {
		t.Errorf("Overwrote key 1: Expected (true, %d), got %v", 17, resp)
	}
}

// Reads a single item from the cache with key="key". Updates the reqID.
func read(key uint64, handle CacheHandle, t *testing.T, reqID *int64) CacheResp {
	handle.ReqQ <- CacheReq{false, key, 0, *reqID}
	*reqID++
	resp := <-handle.RespQ
	if resp.Id != *reqID-1 {
		t.Errorf("Wrong ID. Expected: %d, got %d\n", *reqID-1, resp.Id)
	}
	return resp
}
