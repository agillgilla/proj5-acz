package proj5

/* This is an MNIST classifier service. You can run it by running MnistServer
as a go routine (e.g. "go proj5.MnistServer(handle)") and then passing it
pointers to []byte over handle.reqQ for it to classify. It will send the predicted
label as a string back over handle.respQ.

Note: Due to a bug in one of the underlying libraries (GoLearn), you must have
the environment variable "GODEBUG" set to "cgocheck=0". You can do this by
sourcing env.sh. Do this by calling "source env.sh" (where env.sh is the shell
script included in this repo) before running your program.
*/

import (
	"fmt"
	"strconv"

	"github.com/petar/GoMNIST"
	"github.com/sjwhitworth/golearn/base"
	"github.com/sjwhitworth/golearn/linear_models"
	"gonum.org/v1/gonum/mat"
)

const trainDataPath = "./data/train-images-idx3-ubyte.gz"
const trainLblPath = "./data/train-labels-idx1-ubyte.gz"

type MnistReq struct {
	Val []byte
	Id  int64
}

type MnistResp struct {
	Val int
	Id  int64
	Err error
}

type MnistHandle struct {
	ReqQ  chan MnistReq
	RespQ chan MnistResp
}

/* A classification service. You send it images (as []bytes) and it returns a
label. It works for hand-written digits from the MNIST data set. */
func MnistServer(handle MnistHandle) {

	defer close(handle.RespQ)

	// Load and parse the data from csv files
	rawTrain, err := GoMNIST.ReadSet(trainDataPath, trainLblPath)
	if err != nil {
		panic(fmt.Sprintf("Failed to load training set from %s and %s: %v\n",
			trainDataPath, trainLblPath, err))
	}

	// Create a GoNum matrix from train, converting to float64 along the way
	trainInstance := convertRawToInstance(rawTrain.NRow, rawTrain.NCol+1,
		rawTrain.Images, rawTrain.Labels)

	// Create a new linear SVC with some good default values
	classifier, err := linear_models.NewLinearSVC("l1", "l2", true, 1.0, 1e-4)
	if err != nil {
		panic(err)
	}
	base.Silent()
	classifier.Fit(trainInstance)

	for req := range handle.ReqQ {

		/* Convert the image into a Instance type as required by the classifier.
		* Note, the convertRawToInstance is designed to accept a matrix, so we turn
		* our 1D image into a 2D array with only 1 row. We also provide an empty
		* set for labels as a signal to the function that it can ignore labels */
		testIm := GoMNIST.RawImage(req.Val)
		testInstance := convertRawToInstance(1, rawTrain.NCol+1, []GoMNIST.RawImage{testIm}, []GoMNIST.Label{})

		pred, err := classifier.Predict(testInstance)
		if err != nil {
			// Classifier had a problem, return an error to the client
			handle.RespQ <- MnistResp{0, req.Id, err}
		} else {
			classFloat, err := strconv.ParseFloat(base.GetClass(pred, 0), 64)
			if err != nil {
				// The label returned from our classifier doesn't make sense
				handle.RespQ <- MnistResp{0, req.Id, fmt.Errorf("Couldn't parse classifier output: %v", err)}
			} else {
				// Got a reasonable label from the classifier
				handle.RespQ <- MnistResp{int(classFloat), req.Id, nil}
			}
		}
	}
}

// Converts a raw byte array representing one MNIST digit (without label) into
// an instance that is consumable by the classifier.
func convertRawToInstance(nrow, ncol int, dat []GoMNIST.RawImage, lbl []GoMNIST.Label) *base.Mat64Instances {
	/* While the data comes in as raw bytes, the classifier needs it to be in the
	* form of a base.Mat64Instances. One big difference here is that the
	* classifier is designed to work on batches of images, so it takes matrices
	* as the basic underlying data type. */

	datMat := mat.NewDense(nrow, ncol, nil)

	for rowx := 0; rowx < nrow; rowx++ {
		//Add row to matrix
		for colx := 0; colx < ncol-1; colx++ {
			datMat.Set(rowx, colx, float64(dat[rowx][colx]))
		}

		//Add the label (if labels were included)
		if len(lbl) != 0 {
			datMat.Set(rowx, ncol-1, float64(lbl[rowx]))
		}
	}

	// Turn GoNum matrix into an "Instance" for GoLearn
	datInstance := base.InstancesFromMat64(nrow, ncol, datMat)
	// Mark the last column as the label:
	datInstance.AddClassAttribute(
		base.GetAttributeByName(datInstance, fmt.Sprintf("%d", ncol-1)))

	return datInstance
}
