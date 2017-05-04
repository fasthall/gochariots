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
To quickstart a shared log, use the scripts in test folder in the following order:

1. `sh gochariots/test/controller.sh`
2. `sh gochariots/test/app.sh`
3. `sh gochariots/test/batcher.sh`
4. `sh gochariots/test/filter.sh`
5. `sh gochariots/test/queue.sh`
6. `sh gochariots/test/maintainer.sh`

The above scripts will start a small cluster. The new component information will be reported to the controller. The cluster will have following components:

* **1 controller** running on port **8081**
* **1 app** running on port **8080**
* **3 batchers** running on port **9000**, **9001**, and **9002**
* **1 filter** running on port **9010**
* **3 queues** running on port **9020**, **9021**, and **9022**
* **3 maintainers** running on port **9030**, **9031**, and **9032**

To terminate all the processes, use `sh gochariots/test/kill.sh`.

## Start your own cluster
The cluster needs a controller for meta data.

* `gochariots-controller PORT NUM_DC DC_ID`

The following components can be scaled independently. That is, you can start as many instances as you want for each component. `PORT` is the port the component will listen to, `NUM_DC` is the total number of data centers, and `DC_ID` is the ID number of the data center(start from 0). *Notice that only one queue can hold the token at first.*

* `gochariots-app PORT NUM_DC DC_ID`
* `gochariots-batcher PORT NUM_DC DC_ID`
* `gochariots-filter PORT NUM_DC DC_ID`
* `gochariots-queue PORT NUM_DC DC_ID TOKEN(true|false)`
* `gochariots-maintainer PORT NUM_DC DC_ID`

Remember to report new component to the controller and app.

* `curl -XPOST APP_IP/batcher?host=BATCHER_IP`
* `curl -XPOST CONTROLLER_IP/batcher?host=BATCHER_IP`
* `curl -XPOST CONTROLLER_IP/filter?host=FILTER_IP`
* `curl -XPOST CONTROLLER_IP/queue?host=QUEUE_IP`
* `curl -XPOST CONTROLLER_IP/maintainer?host=MAINTAINER_IP`
* `curl -XPOST CONTROLLER_IP/remote/batcher?dc=DC_ID\&host=BATCHER_IP` (for multi-dc environment)

## Appending record
To append to the shared log, send POST request to http://localhost:8080/record. The payload needs to be in JSON format. See [post_example.sh](test/post_example.sh) and [example.json](test/example.json).

## Issues (TODO)

* Preallocated buffer may not be big enough.
* Indexing hash collision needs to be checked.
* Multiple tag read support.

## Design discussion

* Currently sender(propogation) is bundled with maintainer. This eliminate the communication between sender and maintainer, but can't be scaled. If seperate them, sender can be scaled independently, but require extra communication.
* Does token need to carry all records with unsatisfied dependencies? Currently it does that.
* Batchers should be put behind a load balancer.