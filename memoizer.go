package memoizer

import (
	"github.com/61c-teach/sp18-proj5"
	"hash/crc64"
)

/* The simplest possible implementation that does anything interesting.
This doesn't even do memoization, it just proxies requests between the client
and the classifier. You will need to improve this to use the cache effectively. */
func Memoizer(memHandle proj5.MnistHandle, classHandle proj5.MnistHandle, cacheHandle proj5.CacheHandle) {

    classifierOk := true
    cacheOk := true

	for req := range memHandle.ReqQ {
		crcTable := crc64.MakeTable(crc64.ECMA)
    	reqKey := crc64.Checksum(req.Val, crcTable)
    	//fmt.Printf("Checksum: %x \n", reqKey)
        var cacheResp proj5.CacheResp
        var ok bool

        cacheReadReq := proj5.CacheReq{false, reqKey, 0, req.Id} 
        if cacheOk {
        	cacheHandle.ReqQ <- cacheReadReq
        	cacheResp, ok = <-cacheHandle.RespQ
        }

        if !ok {
            cacheOk = false
        }

    	if cacheOk && cacheResp.Exists { // Request is already in cache, read it from memory
    		if cacheResp.Id == cacheReadReq.Id { // Id is correct (for out of order channel responses)
    			finalResp := proj5.MnistResp{cacheResp.Val, cacheReadReq.Id, nil}
    			memHandle.RespQ <- finalResp
    		}
    	} else { // Request is not in cache, calculate it and memoize it
    		var finalResp proj5.MnistResp
            var ok bool

            if classifierOk {
                classHandle.ReqQ <- req
                finalResp, ok = <-classHandle.RespQ
            }

            if !ok {
                classifierOk = false
            }

            if !classifierOk && !cacheOk {
                finalResp.Err = proj5.CreateMemErr(proj5.MemErr_serCrash, "Classifier and Cache Crashed", nil)
            } else if !classifierOk && cacheOk {
                finalResp.Err = proj5.CreateMemErr(proj5.MemErr_serCrash, "Classifier Crashed, result not in cache", nil)
            } else if classifierOk && !cacheOk {
                // Our cache crashed but our classifier is stil fine.  Don't need to do anything.
            } else if finalResp.Id != req.Id {
    			finalResp.Err = proj5.CreateMemErr(proj5.MemErr_serCorrupt, "Classifier ID Error", nil)
            } else if finalResp.Err != nil {
                finalResp.Err = proj5.CreateMemErr(proj5.GetErrCause(finalResp.Err), "Classifier Error", nil)

    		} else { // No problems with request, caching result
        		cacheWriteReq := proj5.CacheReq{true, reqKey, finalResp.Val, req.Id}
        		cacheHandle.ReqQ <- cacheWriteReq
            }

        		memHandle.RespQ <- finalResp
    	}

    	

		/*classHandle.ReqQ <- req
		resp := <-classHandle.RespQ
		memHandle.RespQ <- resp*/
	}
	close(memHandle.RespQ)
    if cacheOk {
        close(cacheHandle.RespQ)
    }
    if classifierOk {
        close(classHandle.RespQ)
    }
}
