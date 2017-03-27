# Run gochariots in Docker
The image is derived from official golang:1.8-alpine image. Please see [Dockerfile](../Dockerfile) for details.

To deploy a complete gochariots cluster, there are 6 components required. For the functionality of each component, please refer to the [paper](https://openproceedings.org/2015/conf/edbt/paper-236.pdf) or [README.md](../README.md).

To run the container, simply run `docker run --rm --name CONTAINER_NAME -p HOST_PORT:APP_PORT -d fasthall/gochariots COMMAND`

Each component has its own command. Replace `COMMAND` above with one of the following commands.

* gochariots-controller PORT
* gochariots-app PORT
* gochariots-batcher PORT
* gochariots-filter PORT
* gochariots-queue PORT TOKEN(true|false)
* gochariots-maintainer PORT

Notice that `PORT` needs to be consistent to `APP_PORT` in docker command.

If permanent storage is desired, when run `gochariots-maintainer`, use the following command instead.

`docker run --rm --name CONTAINER_NAME -p HOST_PORT:APP_PORT -d -v HOST_DIR:/go/flstore/ fasthall/gochariots gochariots-maintainer APP_PORT`
