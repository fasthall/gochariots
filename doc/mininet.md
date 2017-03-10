# ***WORK IN PROGRESS***

# Mininet simulation

In this document kept all the steps to use mininet to do the simulation. For the introduction to mininet please refer to the [official website](http://mininet.org/).

The following steps were executed on Ubuntu Trusty server. The mininet version is 2.3.0d1, installed from source. The simulation wasn't run on other environments but it should work on most recent distros.

## Installation

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
## Setup

Launch mininet under test/mininet folder:
```
$ cd $GOPATH/src/github.com/fasthall/gochariots/test/mininet
$ sudo mn --custom custom.py --topo single
```
Now we have mininet up and several components running. The next step is report all components info to controller and app. (This will hopefully done automatically in the future version)
```
mininet> init
```
If no error occured, the screen will show several ip and port being added. Now we have a gochariots cluster working. Try `post` and `get` command to append and get records.
```
mininet> post ../example.json
HTTP/1.1 200 OK
Date: Fri, 10 Mar 2017 22:50:04 GMT
Content-Length: 0
Content-Type: text/plain; charset=utf-8

mininet> get 1
{"Causality":{"Host":0,"TOId":0},"Host":0,"LId":1,"TOId":1,"Tags":{"key":"value"}}
``` 

## Configure cluster
To configure cluster, modify [custom.py](../test/mininet/custom.py).