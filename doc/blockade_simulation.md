#  Blockade simulation

Blockade uses Docker containers to run applications and manage network. For details, see the [website](http://blockade.readthedocs.io/).

## Installation

Blockade can be installer natively or runs in VM.

Vagrantfile is provided for Vagrant to start a VM. To use VM to run Blockade, make sure Vagrant and VirtualBox are installed. Clone the [github repo](https://github.com/worstcase/blockade) and run `vagrant up`. It may took a while to start the VM.

    $ git clone https://github.com/worstcase/blockade
    $ cd blockade
    $ vagrant up

Blockade can be installed via `pip` or `easy_install`:

    $ pip install blockade

## Start a cluster

The configuration of the cluster is described in [blockade.yaml](../test/blockade/blockade.yaml). For the explanation of the configuration file, see the [website](http://blockade.readthedocs.io/en/latest/config.html#config).

Use the following commands to start a cluster:

    $ cd test/blockade
    $ blockade up
    $ sh report.sh

The application will export RESTful APIs using port 8080. See [README.md](../README.md) for how to append records.