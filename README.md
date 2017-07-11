# gochariots
Scalable shared log *Chariots* implemented in Go.
For the system details, please refer to the EDBT paper:
[Chariots : A Scalable Shared Log for Data Management in Multi-Datacenter Cloud Environments](https://openproceedings.org/2015/conf/edbt/paper-236.pdf)

## Using Docker
If you are seeing this page on Docker repository page, [please see this page for details](https://github.com/fasthall/gochariots/blob/master/doc/docker.md).

## Quickstart
The shared log system consists of several parts:

* **controller** for meta data and cluster management
* **app** for RESTful APIs
* **batcher** for batching the records sent by applications or propagation
* **filter** for ensuring the uniqueness of records
* **queue** for resolving dependencies
* **maintainer** for physical storage of records

Each part can be scaled independently.
To quickstart a shared log cluster, use the [docker-compose.yaml](deploy/docker-compose.yaml) file in deploy folder.

    $ cd deploy
    $ docker-compose up -d

The above commands will start a small cluster. The new component information will be reported to the controller. The cluster will have following components:

* **1 controller** running on port **8081**
* **1 app** running on port **8080**
* **1 batchers**
* **1 filter**
* **1 queues**
* **1 maintainers** 

To tear down all the resources, run `docker-compose down`.

The internal logs of the system will be mounted as `logs`. The records will be mounter as `flstore`.

Use docker service scale to scale the components.

## Start your own cluster
Modify [docker-compose.yaml](deploy/docker-compose.yaml).
**TO BE UPDATED**

## Appending record
To append to the shared log, send POST request to http://localhost:8080/record. The payload needs to be in JSON format. See [post_example.sh](test/post_example.sh) and [example.json](test/example.json).

## Issues (TODO)

## Design discussion

* Currently sender(propogation) is bundled with maintainer. This eliminate the communication between sender and maintainer, but can't be scaled. If seperate them, sender can be scaled independently, but require extra communication.
* Does token need to carry all records with unsatisfied dependencies? Currently it does that.