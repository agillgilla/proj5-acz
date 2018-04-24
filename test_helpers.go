package proj5

/* This file contains helper functions for testing code in proj5. Feel free
to use these in your own unit tests (the one's we've provided to you use a few
of these).
*/

import (
	"testing"

	"github.com/petar/GoMNIST"
)

/* Tries to classify an image using the classifier service. If "expect" is in
the range (0,9), then it will check if the return label matches "expect",
otherwise it will just make sure the returned label is reasonable. This test
does one image at a time and updates reqID each time.

Returns: the label */
func CheckImage(im []byte, expect int, handle MnistHandle, reqID *int64, t *testing.T) int {
	// Send the image to the classifier
	handle.ReqQ <- MnistReq{im, *reqID}
	defer func() {
		*reqID++
	}()

	// Block waiting for a response
	resp, ok := <-handle.RespQ
	if !ok {
		t.Error("Channel closed unexpectedly. Did the classifier die?\n")
	}

	if resp.Id != *reqID {
		t.Errorf("Classifier provided wrong ID. Expected %d, got %d\n", *reqID, resp.Id)
	}

	if resp.Err != nil {
		t.Errorf("Classifier internal error: %v", resp.Err)
	}

	if expect < 0 || expect > 9 {
		// No "expect" provided, just check if the label makes sense
		if resp.Val < 0 || resp.Val > 9 {
			t.Errorf("Unreasonable classification: %d", resp.Val)
		}
	} else {
		if resp.Val != expect {
			t.Errorf("Expected %d got %d\n", expect, resp.Val)
		}
	}
	return resp.Val
}

/* Checks many images at once (can take advantage of buffered channels)
Check if the classifier correctly classifies the test images.
if "expect" is nil, then we only check that the class is
reasonable without looking at it's exact value. This is useful because the
classifier isn't 100% accurate, we just want to make sure it's behaving in a
reasonable way.
Returns: Classes (even if therere were errors or they weren't reasonable)
*/
func CheckImages(ims []GoMNIST.RawImage, expects []int, handle MnistHandle, reqID *int64, t *testing.T) []int {
	// The index of the next thing to send
	reqX := 0

	// Holds all the responses regardless of failures
	resps := make([]int, len(ims))

	/* Count how many responses we've received so we know when to stop waiting for them */
	respCount := 0

	//The reqId we started with
	initReqId := *reqID
	// Update the reqID at the end
	defer func() {
		*reqID += int64(len(ims))
	}()

	// Send requests and drain responses as fast as possible
	for reqX < len(ims) {
		// The select will keep sending requests until the channel fills up, or
		// some responses come back.
		// The select will do whichever case it can. If we didn't have this and the
		// reqQ filled up, we could deadlock because the classifier would be
		// blocked trying to send responses, but we would never see the responses
		// because we were blocked trying to send a new request.
		select {
		// Send images to the memoizer in order, use the index offset by
		// initReqID as the request ID (offsetting keeps it globally unique
		// within a test)
		case handle.ReqQ <- MnistReq{ims[reqX], int64(reqX) + initReqId}:
			// We were able to send a request, move on to trying to send the next request
			reqX++
		case resp, ok := <-handle.RespQ:
			if !ok {
				t.Error("Channel closed unexpectedly. Did the classifier die?\n")
			}

			// Index in ims/expect that this response corresponds to
			respX := resp.Id - initReqId
			// A response came. We know which image it was by the resp.Id
			if expects == nil {
				if resp.Val < 0 || resp.Val > 9 {
					t.Errorf("Unreasonable classification: %d", resp.Val)
				}
			} else {
				if resp.Val != expects[respX] {
					t.Errorf("Expected %d got %d\n", expects[respX], resp.Val)
				}
			}
			resps[respX] = resp.Val
			respCount++
		}
	}

	// Keep waiting for responses until we've received them all
	for respCount < len(ims) {
		resp, ok := <-handle.RespQ
		if !ok {
			t.Error("Channel closed unexpectedly. Did the classifier die?\n")
		}

		// Index in ims/expect that this response corresponds to
		respX := resp.Id - initReqId

		if expects == nil {
			if resp.Val < 0 || resp.Val > 9 {
				t.Errorf("Unreasonable classification: %d", resp.Val)
			}
		} else {
			if resp.Val != expects[respX] {
				t.Errorf("Expected %d got %d\n", expects[respX], resp.Val)
			}
		}
		resps[respX] = resp.Val
		respCount++
	}

	return resps
}
