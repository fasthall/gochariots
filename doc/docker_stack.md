# Running a shared log using Docker

In this document, step-by-step instructions to use Docker to run a GoChariots log cluster are provided.

## Using Docker Compose

### Prerequisite

#### Docker Engine
A machine with Docker 1.13.0+ installed. 17.06.0+ recommended. To install Docker, please refer to [this page](https://docs.docker.com/engine/installation/).

#### Docker Compose
Install docker-compose via pip. It can be done by `pip install docker-compose`. If other installation method is desired, please refer to [this page](https://docs.docker.com/compose/install/).

### Start a GoChariots log
Make sure there are two files under your working directory:
* docker-compose.yaml
* config/config.yaml

These files can be found in [the deploy folder in the repo](https://github.com/fasthall/gochariots/tree/master/deploy).

Use *docker-compose* to start a cluster.

    $ docker-compose up -d

You will see some messages showing that Docker is creating resources.

    Creating network "deploy_cluster" with the default driver
    Creating deploy_controller_1 ... 
    Creating deploy_controller_1 ... done
    Creating deploy_indexer_1 ... 
    Creating deploy_batcher_1 ... 
    Creating deploy_filter_1 ... 
    Creating deploy_queue-leader_1 ... 
    Creating deploy_app_1 ... 
    Creating deploy_maintainer_1 ... 
    Creating deploy_queue_1 ... 
    Creating deploy_queue-leader_1
    Creating deploy_batcher_1
    Creating deploy_filter_1
    Creating deploy_indexer_1
    Creating deploy_maintainer_1
    Creating deploy_app_1
    Creating deploy_maintainer_1 ... done
    Creating deploy_queue_1 ... done
    Creating deploy_batcher_lb_1 ... done

Use browser or curl to see the output of `localhost:8081`. You should see something like this:

    {"apps":["172.21.0.7:8080"],"batchers":["172.21.0.6:9000"],"filters":["172.21.0.4:9010"],"indexers":["172.21.0.5:9040"],"maintainers":["172.21.0.8:9030"],"queues":["172.21.0.3:9020","172.21.0.9:9020"],"remoteBatchers":null}

Make sure you have running instance for each component except remoteBatchers. The current version of GoChariots doesn't check the healthy of the system, so make sure the controller has all the information before you post a record.

Now the system is running. You can post a record to the shared log by using HTTP Post or by sending raw data to batcher via TCP socket. Try

    $ curl -i -XPOST -H "Content-Type: application/json" -d '{"tags": {"key": "value"}, "prehash": 0, "seed": 0}' http://localhost:8080/record

### Scaling
To scale a stage, use `--scale` option along with `docker-compose up`.

    $ docker-compose up -d --scale batcher=3

Now type `docker ps`, you should see another two new batchers running:

    CONTAINER ID        IMAGE                 COMMAND                  CREATED             STATUS              PORTS                                               NAMES
    c3ac364fc012        fasthall/gochariots   "gochariots batche..."   21 seconds ago      Up 19 seconds       0.0.0.0:32807->9000/tcp                             deploy_batcher_3
    2e93b7b7dc63        fasthall/gochariots   "gochariots batche..."   21 seconds ago      Up 19 seconds       0.0.0.0:32806->9000/tcp                             deploy_batcher_2
    ...

The batchers are load balanced by HAProxy. User can directly access to any of the batchers via port 9000.

### Tear down
To tear down the cluster, simply type

    $ docker-compose down

## Using Docker Swarm
Docker Swarm is a cluster management integrated with Docker Engine.

### Prerequisite

#### Docker Engine
A machine with Docker 1.13.0+ installed. 17.06.0+ recommended. To install Docker, please refer to [this page](https://docs.docker.com/engine/installation/).

### Open ports
Make sure the following ports are open.

* TCP port 2377 for cluster management communications
* TCP and UDP port 7946 for communication among nodes
* UDP port 4789 for overlay network traffic

#### Run Docker in Swarm mode
In this example, we have 3 machines with Docker Engine installed: 1 manager, 2 workers.

Firstly, run the following command on manager:

    $ docker swarm init --advertise-addr <MANAGER-IP>

You should see message like this:

    Swarm initialized: current node (90vi90f06s26hjaz984tbx1ws) is now a manager.

    To add a worker to this swarm, run the following command:

        docker swarm join --token SWMTKN-1-48rtut1addvgqx6lfulb0agzdtd379o6hokzu6flbjwtj9bun4-c3qknm8hqvtalx8ygpl10ssxr 192.168.65.2:2377

    To add a manager to this swarm, run 'docker swarm join-token manager' and follow the instructions.

Run the `docker swarm join` command showing above on other workers. Once done, check the node list by typing `docker node ls` on manager.

    ID                            HOSTNAME            STATUS              AVAILABILITY        MANAGER STATUS
    98jntg1no623w2ik8k6u79cko     euca-10-1-4-17      Ready               Active              
    iozia9dgkmj3exfgo5kawavgr     euca-10-1-5-208     Ready               Active              
    xof5xqosvt8wtdbcw5hdfnmiv *   euca-10-1-4-206     Ready               Active              Leader

Now we have three nodes in the swarm. The next step is to prepare the required config files and storage space for each node. From [this folder](https://github.com/fasthall/gochariots/tree/master/deploy), copy `config/config.yaml` to **all** the node machines. Also, make new directories `logs` and `flstore` on **all** the node machines. Your working directory of **all** the node machines should look like this:

    $ pwd
    /home/ubuntu/deploy
    $ find .
    .
    ./logs
    ./config
    ./config/config.yaml
    ./flstore

If your working directory is not `/home/ubuntu/deploy`, make sure you modify the path in [docker-stack.yaml](https://github.com/fasthall/gochariots/blob/master/deploy/docker-stack.yaml). Put docker-stack.yaml in the working directory.

On your manager machine with [docker-stack.yaml](https://github.com/fasthall/gochariots/blob/master/deploy/docker-stack.yaml) prepared, run the following command:

    docker stack deploy -c docker-stack.yaml gochariots

Wait for swarm to create services for few seconds. Now there will be several services running.

    $ docker stack ps gochariots
    z08v8f3i54jf        gochariots_indexer.1        fasthall/gochariots:latest   euca-10-1-5-208     Running             Running 33 seconds ago                       
    fhlmkcmk00j0        gochariots_queue-leader.1   fasthall/gochariots:latest   euca-10-1-4-206     Running             Running 17 seconds ago                       
    rcoftubuisg9        gochariots_controller.1     fasthall/gochariots:latest   euca-10-1-5-208     Running             Running 36 seconds ago                       
    w0qomy1zw1x0        gochariots_batcher.1        fasthall/gochariots:latest   euca-10-1-4-17      Running             Running 39 seconds ago                       
    3017lbu9dn09        gochariots_queue.1          fasthall/gochariots:latest   euca-10-1-4-17      Running             Running 41 seconds ago                       
    pn3ndgdnegim        gochariots_maintainer.1     fasthall/gochariots:latest   euca-10-1-4-206     Running             Running 13 seconds ago                       
    02ab34ao1lz0        gochariots_filter.1         fasthall/gochariots:latest   euca-10-1-4-17      Running             Running 45 seconds ago                       
    n5czqspb90ym        gochariots_app.1            fasthall/gochariots:latest   euca-10-1-5-208     Running             Running 46 seconds ago                       
    p0168jz2kefl        gochariots_batcher_lb.1     dockercloud/haproxy:latest   euca-10-1-4-206     Running             Running 19 seconds ago                       
    i3vd6lee55dp        gochariots_batcher.2        fasthall/gochariots:latest   euca-10-1-4-206     Running             Running 12 seconds ago                       
    sumi5ioz1evs        gochariots_batcher.3        fasthall/gochariots:latest   euca-10-1-5-208     Running             Running 39 seconds ago

#### Scaling
To scale the service, type

    docker service scale gochariots_batcher=5

Like Docker Compose, the batchers are load balanced and can be accessed by port 9000 of any of the nodes.

#### Tear down

To tear down the swarm stack, use

    $ docker stack rm gochariots

## Post records

### HTTP Post

Post JSON file to app's 8080 port. For example, 

    $ curl -i -XPOST -H "Content-Type: application/json" -d '{"tags": {"key": "value"}, "prehash": 0, "seed": 0}' http://localhost:8080/record

The JSON file needs three fields: `tags`, `prehash`, and `seed`.

#### Tags
Tags is a key-value pair map. There is no limitation on the content of key and value. Application can put any of useful information in the pair.

#### Seed and Hash
Seed and prehash are used to specify the casual dependency. All of the events in the same event chain should have the same seed. For example, consider the following event chain:

    func A(k1, v1) -> func B(k2, v2)

Each function invocation has its own record, named record A and record B. The arguments are the tags' content.

Both record A and record B should share a same seed. The seed can be assigned arbitrarily, as long as it's a 64bit integer. To make sure record B is casually dependent on record A, we need to get the fnv-1a hash of `k1:v1`, 9489172654252433228, and assign this value as record B's prehash. All the provided libraries have function to calculate the hash.

### TCP stream
Use can also opt to directly send the data to batcher. To achieve that, simply establish a TCP connection to any of batchers, and send the data.

The data should follow the following format:

* The first 4 bits indicate the length of this packet(big endian).
* The 5th bit should be character 'r'(114).
* The remaining data is a JSON object.

The JSON object is slightly different with the JSON object we used to post record via HTTP post, but the context and meaning of fields is same. This is an example of JSON object:

    {"Host":0,"Tags":{"key:value"},"Pre":{"Hash":0},"Seed":0}

Refer to [rawpost.py](../test/script/rawpost.py) for python implement.

### Libraries
Libraries for [Python](https://github.com/fasthall/gochariots-python-lib), [Node.js](https://github.com/fasthall/gochariots-nodejs-lib), and [Java](https://github.com/fasthall/gochariots-java-lib) are provided.