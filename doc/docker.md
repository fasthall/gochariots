# Run gochariots in Docker
The image is derived from official golang:1.8-alpine image. Please see [Dockerfile](../Dockerfile) for details.

To deploy a complete gochariots cluster, there are 6 components required. For the functionality of each component, please refer to the [paper](https://openproceedings.org/2015/conf/edbt/paper-236.pdf) or [README.md](../README.md).

To run the container, simply run `docker run --rm --name CONTAINER_NAME -p HOST_PORT:APP_PORT -d fasthall/gochariots gochariots COMMAND`

Each component has its own command. Replace `COMMAND` above with one of the following commands.

* app [<flags>] <port> <num_dc> <id>
* batcher [<flags>] <port> <num_dc> <id>
* controller [<flags>] <port> <num_dc> <id>
* filter [<flags>] <port> <num_dc> <id>
* queue --[no-]hold [<flags>] <port> <num_dc> <id>
* maintainer [<flags>] <port> <num_dc> <id> [<ntime>]
* indexer [<flags>] <port> <num_dc> <id>

Notice that `port` needs to be consistent to `APP_PORT` in docker command.

If permanent storage is desired, when run `gochariots maintainer`, add the following option: `-v HOST_DIR:/go/flstore/`.
