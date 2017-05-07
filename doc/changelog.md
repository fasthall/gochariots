## May 5, 2017
### Batcher supports reading multiple records at once
This is to verify that the reason batcher is the bottleneck of whole system is that it takes too much time to decode JSON(decoding a single large JSON is faster then decoding several small JSON files).

---
## May 4, 2017
### Use GOB to transfer data across stages instead of JSON
Golang's JSON parser implementaion is very slow. The likely reason is the parsing is dynamic. There exist [tools](https://github.com/pquerna/ffjson) to accelerate JSON parser, but all need to be used as pre-compiling tool.

GOB is golang's proprietary protocol. It's 2~3 times faster than JSON. The problem is that it can't be used by other languages.

The current implementation uses GOB to transfer data across stages, except batchers only read JSON input(so clients can be implemented with different languages).

---
## May 3, 2017
### Add indexer, bundled with maintainer.
Indexer is responsible to indexing the records. If a client wants to get a record by LId, it sends the request to maintainer directly. However, if the client only have tags information, it needs to ask indexer for the LId.

The current implementation is maintaining a in-memory lookup table for record tags. The table uses the [fnv hash](https://golang.org/pkg/hash/fnv/) of the tags as key, and the LId as value.

### Maintainer writes to a single file
The permanent storage format needs to be discussed. For the simplicity of implementation, currently all records are written to a single divided into 512-byte blocks.