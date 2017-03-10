#  Blockade simulation

Blockade uses Docker containers to run applications and manage network. For details, see the [website](http://blockade.readthedocs.io/).

## Installation
Blockade can be installer natively or runs in VM. For the sake of simplicity, [Vagrant](https://www.vagrantup.com/) and [VirtualBox](https://www.virtualbox.org/) are used in this document. Please follow the [instruction](https://www.vagrantup.com/docs/installation/) to install Vagrant and VirtualBox.

After Vagrant and VirtualBox are installed, clone the [github repo](https://github.com/worstcase/blockade) and `vagrant up`. It may took a while to start the VM.

    $ git clone https://github.com/worstcase/blockade
    $ cd blockade
    $ vagrant up