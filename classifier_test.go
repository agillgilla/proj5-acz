package proj5

import (
	"fmt"
	"testing"

	"github.com/petar/GoMNIST"
)

func TestMnist(t *testing.T) {
	rawTrain, err := GoMNIST.ReadSet(trainDataPath, trainLblPath)
	if err != nil {
		panic(fmt.Sprintf("Failed to load training set from %s and %s: %v\n",
			trainDataPath, trainLblPath, err))
	}

	/* Create the channels that we will use for communicating with the classifier */
	handle := MnistHandle{
		make(chan MnistReq, 100),
		make(chan MnistResp, 100),
	}
	/* Start the MnistServer as a goroutine. It will execute concurrently with the
	main thread from now on. */
	go MnistServer(handle)

	// Request ID increments once per message to stay unique
	var reqID int64 = 0

	// == Basic classification test ==
	// Keep in mind, not all inputs must give the right label because this
	// classifier isn't perfect. In fact, offline tests show that it's only about
	// 86% accurate. I've verified manually that the classifier gets this one
	// right though. If you change the classifier, and this starts failing, check
	// to make sure it's at least reasonable.
	t.Run("Basic Correctness:", func(t *testing.T) {
		CheckImage(rawTrain.Images[0], int(rawTrain.Labels[0]), handle, &reqID, t)

		// == Make sure it works the same way twice in a row == */
		CheckImage(rawTrain.Images[0], int(rawTrain.Labels[0]), handle, &reqID, t)

		// == Test some more values, make sure the model is deterministic.
		//These tests should always pass, no matter the model since they only check
		//that the response is reasonable
		firstResp := CheckImage(rawTrain.Images[1], -1, handle, &reqID, t)
		if resp := CheckImage(rawTrain.Images[1], -1, handle, &reqID, t); resp != firstResp {
			t.Errorf("Classification on second attempt doesnt match first, %d != %d", firstResp, resp)
		}
	})

	t.Run("Batch Test:", func(t *testing.T) {
		// == Test a bunch of images to make sure the model doesn't crash under load
		CheckImages(rawTrain.Images, nil, handle, &reqID, t)
	})

	// == Close the channel (this must be the last test!)
	close(handle.ReqQ)
}
