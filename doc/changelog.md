# June

## June 21, 2017
### Use BoltDB to store indexes in indexer
By default, all the indexes(hash and seed pair) are stored in memory. An instance with 4GB memory can roughly store 500 million entries. If exceeded, there's no overflow handling currently implemented. Using *-db* option to launch indexer will tell the indexer to use BoltDB and go-cache to access indexes. Each will be in the cache for 5 minutes then expire. The BoltDB file is stored as `indexes.db`. *Notice that this implementation is very slow, not optimized now*

## June 18, 2017
### Merge pretoid and no_deferred branches
There are only master and pretoid branches now. In master branch we use hash, while in pretoid branch we use toid. By default, token doesn't carry deferred records. To carry deferred records with token, use *-c* option, for example: `gochariots-queue -c 9020 1 0 true`.

## June 16, 2017
### Rewrite system logging using logrus
Rewrite system logging using [logrus](https://github.com/sirupsen/logrus), use *-v* option to turn on all logging infos. For example, `gochariots-app 8080 1 0` by default only write `WARNING` and `ERROR` level events into logging file. `gochariots-app -v 8080 1 0` will turn on all events of logging.

## June 8, 2017
### Compare the performance between carrying v.s. not carrying deferred records with token
Not carrying deferred records with token may cause additional delay, but it's negligible. See [0606.md](experiment/0606.md).

### Use hash and seed to describe causal dependency
A record now has Seed and Hash field. Seed is a randomly generated 64 bits number, or application can give it context and manually assign it to the record. All the events(records) in the same event chain will share an identical seed. Hash is used to specify causal dependency(see changelog of May 15), and serves as the key of indexer.

When a record is stored in maintainer, indexer will hash the content and use the hash as key to store record's seed as value. Now, the deferred record can query indexer whether its prerequisite has been indexed.

# May

## May 18, 2017
### Refactor connection reading function
As in [connection.go](../misc/connection/connection.go).

### Use hash to specify casual dependency instead of tags
Now, record has a field pre:hash which is a uint64 type. To specify casual dependency, firstly hash "key:value" into a 64 bits int using fnv-1a, then put this hash value into the field. The deferred record carried by token will be checked, if the hash value has been indexed by indexers, the queue will remove the record from the token and send it to maintainer. 

## May 15, 2017
### Use hash to specify casual dependency
Client can use [fnv-1a](https://en.wikipedia.org/wiki/Fowler%E2%80%93Noll%E2%80%93Vo_hash_function) hash to specify casual dependency now.
See [event_chain.py](../test/script/event_chain.py) for detail.
### Bug fixed
Queue will send LId query to indexer, not maintainer now.

## May 14, 2017
### Seperate indexer from maintainer
Since the maintainer has more and more work to do, indexer is seperated from maintainer for better performance now.

## May 10, 2017
### Indexing broadcasting
The indexer broadcasts hash instead of complete tags now.
### Bug fixed
Fixed a major bug that tcp read incompletely.

## May 8, 2017
### Indexing broadcasting
Now the indexder will broadcast the tags and corresponding record LId to the subscribers. To subscribe, clients need to send a message to the maintainer. This is for testing blocking event chain v.s. appending non-blocking causal order event. 

## May 5, 2017
### Batcher supports reading multiple records at once
This is to verify that the reason batcher is the bottleneck of whole system is that it takes too much time to decode JSON(decoding a single large JSON is faster then decoding several small JSON files).

## May 4, 2017
### Use GOB to transfer data across stages instead of JSON
Golang's JSON parser implementaion is very slow. The likely reason is the parsing is dynamic. There exist [tools](https://github.com/pquerna/ffjson) to accelerate JSON parser, but all need to be used as pre-compiling tool.

GOB is golang's proprietary protocol. It's 2~3 times faster than JSON. The problem is that it can't be used by other languages.

The current implementation uses GOB to transfer data across stages, except batchers only read JSON input(so clients can be implemented with different languages).

## May 3, 2017
### Add indexer, bundled with maintainer.
Indexer is responsible to indexing the records. If a client wants to get a record by LId, it sends the request to maintainer directly. However, if the client only have tags information, it needs to ask indexer for the LId.

The current implementation is maintaining a in-memory lookup table for record tags. The table uses the [fnv hash](https://golang.org/pkg/hash/fnv/) of the tags as key, and the LId as value.

### Maintainer writes to a single file
The permanent storage format needs to be discussed. For the simplicity of implementation, currently all records are written to a single divided into 512-byte blocks.