# Mininet simulation

In this document kept all the steps to use mininet to do the simulation. For the introduction to mininet please refer to the [official website](http://mininet.org/).

The following steps were executed on Ubuntu Trusty server. The mininet version is 2.3.0d1, installed from source. The simulation wasn't run on other environments but it should work on most recent distros.

## Installation

Install latest Docker engine.

    $ curl -sSL https://get.docker.com/ | sh

(Optional) After installation, add user to docker permission group.

Install mininet and dependencies.

    $ git clone git://github.com/mininet/mininet
    $ cd mininet
    $ git checkout -b 2.3.0d1 (Optional, the latest branch should work)
    $ cd ..
    $ mininet/util/install.sh -a

## Setup
