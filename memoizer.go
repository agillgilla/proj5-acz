package memoizer

import (
	"github.com/61c-teach/sp18-proj5"
)

/* The simplest possible implementation that does anything interesting.
This doesn't even do memoization, it just proxies requests between the client
and the classifier. You will need to improve this to use the cache effectively. */
func Memoizer(memHandle proj5.MnistHandle, classHandle proj5.MnistHandle, cacheHandle proj5.CacheHandle) {

	for req := range memHandle.ReqQ {
		classHandle.ReqQ <- req
		resp := <-classHandle.RespQ
		memHandle.RespQ <- resp
	}
}
