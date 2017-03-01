# gochariots
Scalable shared log *Chariots* implemented in Go.
For the system details, please refer to the EDBT paper:
[Chariots : A Scalable Shared Log for Data Management in Multi-Datacenter Cloud Environments](https://openproceedings.org/2015/conf/edbt/paper-236.pdf)

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

## Appending record
To append to the shared log, send POST request to http://localhost:8080/record. The payload needs to be in JSON format. See [post_example.sh](test/post_example.sh) and [example.json](test/example.json).
