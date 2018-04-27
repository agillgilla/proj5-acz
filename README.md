# sp18-proj5
Services and utilities for proj5. You can use them by adding the following to
your go program (assuming you've cloned this repo using "go get")

# Installation
The best way to use these files is to clone this repo using the "go get"
command. This command will automatically create the necessary directories and
clone the repo. To use go get, run the following command anywhere in your go
workspace:

    $ go get github.com/61c-teach/sp18-proj5

To use in your project, you will need to import it by adding the following to
the beginning of your go program:

    import "github.com/61c-teach/sp18-proj5"

You can use the exported contents (functions, types, and feilds that start with
a capital letter) by prefixing the name with "proj5.". For example, to create a
new CacheReq (a request to read/write to the cache), you could do:

    req := proj5.CacheReq{true, 3, 0, 1}

# classifier.go
The classifier service is designed to translate hand-written digits from the
classic MNIST dataset. The input to it is a []byte (byte slice, see the go
tutorial on slices) which represents a 28x28 pixel image (784 pixels total). If
you're curious, you can use the provided "Show" function in memoizer\_test.go to
save these images as pngs (like the example above). The classifier then runs an
ensemble of support vector machine classifiers, one for each digit, and picks
the most likely digit to return (as an int). Like every service in this
project, the classifier accepts a request ID with every request, and returns
that ID with the corresponding response. This allows for potentially
out-of-order messages (which can occur due to network issues or from
optimizations). Note that the current classifier implementation will always
respond in order, but you are free to return messages out of order from your
memoization service. Note also that the classifier response includes a
potential error message. Under normal operation, this feild is always "nil",
but it may contain an error message if something went wrong. You may not assume
that all classification requests will succeed. Take a look at
classifier\_test.go for examples on how to use the classifier service.

# cache.go
The caching service is essentially a remote hash table. You give it key-value
pairs and it saves them for later. You can then ask it if it's seen a
particular key and it will return the corresponding value (if any). Caching
services like this are extremely common in datacenters, examples include
memcached and redis. Caching services reduce the load on other services by
saving frequent or recent results. Our cache in this project is a very simple
key-value store. You may assume that it will never fill up (no need to deal
with eviction/replacement). You send it requests using the CacheReq struct
which contains a read/write flag, a 64-bit key, an int value, and (like
everything) a 64-bit requestID. If the request is a write, then the cache will
not respond. If the request is a read, then the cache will respond with a
CacheResp struct that contains an "exists" flag (true if the item was found)
and the corresponding value (if any), and the requestID. Check out
cache\_test.go for examples on how to use it.

Note that the key is a 64-bit integer, but the images we are using are byte
slices. To deal with this, you will need to use a hash of the image as a key.
While there could, technically, be collisions on this hash function, the
chances of that are sufficently low for our application to work (you may assume
that collisions will never happen). Afterall, we are using machine learning to
classify the images which is far more likely to give a bad label than a
hash-collision would. We recommend taking a look at Go's crc64 hash function,
although you are free to use any hash function you wish (so long as it has a
similarly low collision probability). It's not recommended to write your own
hash function, there are existing libraries for that sort of thing.
