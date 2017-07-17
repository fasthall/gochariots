# gochariots
Scalable shared log *Chariots* implemented in Go.
For the system details, please refer to the EDBT paper:
[Chariots : A Scalable Shared Log for Data Management in Multi-Datacenter Cloud Environments](https://openproceedings.org/2015/conf/edbt/paper-236.pdf)

## Using Docker
If you are seeing this page on Docker repository page, [please see this page for details](https://github.com/fasthall/gochariots/blob/master/doc/docker.md).

## Quickstart
We can use Docker Compose to run a local shared log, or Docker Swarm to run a cluster on multiple hosts. Please see [this document](doc/docker_stack.md) for detail.

## Appending record
To append to the shared log, send POST request to http://localhost:8080/record. The payload needs to be in JSON format. See [post_example.sh](test/post_example.sh) and [example.json](test/example.json).

## Issues (TODO)
When a component of any stage stops working and the controller couldn't connect to it, the controller should prompt a warning message rather than just panic. 

## Design discussion

* Currently sender(propogation) is bundled with maintainer. This eliminate the communication between sender and maintainer, but can't be scaled. If seperate them, sender can be scaled independently, but require extra communication.
* Does token need to carry all records with unsatisfied dependencies? Currently it does that.