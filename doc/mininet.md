# Mininet simulation

In this document kept all the steps to use mininet to do the simulation. For the introduction to mininet please refer to the [official website](http://mininet.org/).

The following steps were executed on Ubuntu Trusty server. The mininet version is 2.3.0d1, installed from source. The simulation wasn't run on other environments but it should work on most recent distros.

## Installation

### Install from source
Install gochariots binary. Install [Go](https://golang.org/doc/install#install) if it's not installed yet.
```
$ go get github.com/fasthall/gochariots/cmd/...
```
Make sure you have binary executables. Try command `gochariots-controller`. Use `which gochariots-controller` to see where the binary was placed, it should be under `$GOPATH/bin`. Remember this file path.

Install mininet and dependencies.
```
$ git clone git://github.com/mininet/mininet
$ cd mininet
$ git checkout -b 2.3.0d1 (Optional, the latest branch should work)
$ cd ..
$ mininet/util/install.sh -a
```

### Using Vagrant
Another way to do this is using [Vagrantfile](../test/mininet/Vagrantfile) provided in this repo. 
```
$ cd gochariots/test/mininet/
$ vagrant up
```

## Setup

Launch mininet under test/mininet folder:
```
$ cd $GOPATH/src/github.com/fasthall/gochariots/test/mininet
$ sudo mn --link tc --custom custom.py --topo double
```
Now we have mininet up and several components running. The next step is report all components info to controller and app. (This will hopefully done automatically in the future version)
```
mininet> init double
```
If no error occured, the screen will show several ip and port being added. Now we have a gochariots cluster working. Try `post` and `get` command to append and get records. Use `help post` and `help get` to get information.
```
mininet> post 0 ../example.json
HTTP/1.1 200 OK
Date: Fri, 10 Mar 2017 22:50:04 GMT
Content-Length: 0
Content-Type: text/plain; charset=utf-8

mininet> get 0 1
{"Causality":{"Host":0,"TOId":0},"Host":0,"LId":1,"TOId":1,"Tags":{"key":"value"}}
```

To randomly post records to multiple data centers, modify [custom.py](../test/mininet/custom.py) line 272. The usage is:
```
mininet> random_post num_records dependency_prob max_window margin
```
`max_window` and `margin` need to be set carefully, otherwise the records may not be able to be appended due to dependency issue. For example, when `max_window = 3` and `margin = 5`, the 100th record may depend on record 92 to 94. 

### Topology
To run the simulation on different topology, replace single with another topology, including `mn` and `init` command's parameter. For the description of topologies, please see [topology.md](topology.md).

## Configure cluster
To configure cluster, modify [custom.py](../test/mininet/custom.py).
