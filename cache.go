package proj5

// A cache request (read/write) message
type CacheReq struct {
	// True if this is a write, false for reads
	Write bool
	// The key to lookup/write
	Key uint64
	// The value to write if "write==true", don't care otherwise
	Val int
	// A unique identifier for this request, will be returned in the
	// corresponding response
	Id int64
}

// A response from the cache (contains value if found)
type CacheResp struct {
	// True if the item was found, false otherwise
	Exists bool
	// The value that was found (or don't care if exists==false)
	Val int
	// Corresponds to the ID provided in the request
	Id int64
}

// Packages up the communication channels with a cache server */
type CacheHandle struct {
	ReqQ  chan CacheReq
	RespQ chan CacheResp
}

/* A cache server, typically launched as a go routine. You can send key/value
   pairs to it and it will remember them for later. */
func Cache(handle CacheHandle) {

	//Cache should close respQ if it closes or crashes
	defer close(handle.RespQ)

	// The actual cache is just a simple hash-table (aka "map")
	cache := make(map[uint64]int)

	// Loop until the client closes our channel
	for req := range handle.ReqQ {
		// req, closed := <-handle.ReqQ
		// if !closed {
		// 	break
		// }

		if req.Write {
			// Write request
			cache[req.Key] = req.Val
		} else {
			// Read request
			val, ok := cache[req.Key]
			if !ok {
				// Item not found (no such key)
				handle.RespQ <- CacheResp{false, 0, req.Id}
			} else {
				// Item found
				handle.RespQ <- CacheResp{true, val, req.Id}
			}
		}
	}
}
