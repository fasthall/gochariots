In this file, all the topologies in both scrpts folder and mininet custom file are described.

## Simple
A single cluster setup.
* **1 app** on 8080
* **1 controller** on 8081
* **3 batchers** on 9000, 9001, 9002
* **1 filter** on 9010
* **3 queues** on 9020, 9021, 9022
* **3 maintainers** on 9030, 9031, 9032

## Single
A single cluster setup.
* **1 app** on 8080
* **1 controller** on 8081
* **1 batcher** on 9000
* **1 filter** on 9010
* **1 queue** on 9020
* **1 maintainer** on 9030

## Double
Two clusters setup.
### Cluster A(ID=0)
* **1 app** on 8080
* **1 controller** on 8081
* **1 batcher** on 9000
* **1 filter** on 9010
* **1 queue** on 9020
* **1 maintainer** on 9030
### Cluster B(ID=1)
* **1 app** on 8180
* **1 controller** on 8181
* **1 batcher** on 9100
* **1 filter** on 9110
* **1 queue** on 9120
* **1 maintainer** on 9130