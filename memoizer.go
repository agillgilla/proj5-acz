package memoizer

import (
	"github.com/61c-teach/sp18-proj5"
	"hash/crc64"
	"fmt"
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

    	cacheReadReq := proj5.CacheReq{false, reqKey, 0, req.Id} 
    	cacheHandle.ReqQ <- cacheReadReq
    	
    	cacheResp, ok := <-cacheHandle.RespQ

        if !ok {
            cacheOk = false
        }

    	if cacheOk && cacheResp.Exists { // Request is already in cache, read it from memory
    		if cacheResp.Id == cacheReadReq.Id { // Id is correct (for out of order channel responses)
    			finalResp := proj5.MnistResp{cacheResp.Val, cacheReadReq.Id, nil}
    			memHandle.RespQ <- finalResp
    		} else { // ERROR!
    			fmt.Printf("%s", "ERROR! Cache Response ID doesn't match Cache Request ID!")
    		}
    	} else { // Request is not in cache, calculate it and memoize it
    		classHandle.ReqQ <- req
    		finalResp := <-classHandle.RespQ

    		if finalResp.Id != req.Id {
    			finalResp.Err = proj5.CreateMemErr(proj5.MemErr_serCorrupt, "Classifier Error", finalResp.Err)
    		}

    		cacheWriteReq := proj5.CacheReq{true, reqKey, finalResp.Val, req.Id}
    		cacheHandle.ReqQ <- cacheWriteReq

    		memHandle.RespQ <- finalResp
    	}

    	

		/*classHandle.ReqQ <- req
		resp := <-classHandle.RespQ
		memHandle.RespQ <- resp*/
	}
	close(memHandle.RespQ)
	/*close(cacheHandle.RespQ)
	close(classHandle.RespQ)*/
}
