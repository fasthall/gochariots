import os
import time
from mininet.topo import Topo
from mininet.cli import CLI

class Single(Topo):
    def __init__(self):
        Topo.__init__(self)
        
        # Add hosts and switches
        controller = self.addHost('c')
        app = self.addHost('a')
        batcher = self.addHost('b1')
        filter= self.addHost('f1')
        queue = self.addHost('q1')
        maintainer = self.addHost('m1')
        switch = self.addSwitch('s1')
        client = self.addHost('client')

        # Add links
        self.addLink(controller, switch)
        self.addLink(app, switch)
        self.addLink(batcher, switch)
        self.addLink(filter, switch)
        self.addLink(queue, switch)
        self.addLink(maintainer, switch)
        self.addLink(client, switch)

class Simple(Topo):
    def __init__(self):
        Topo.__init__(self)

        # Add hosts and switches
        controller = self.addHost('c')
        app = self.addHost('a')
        batcher1 = self.addHost('b1')
        batcher2 = self.addHost('b2')
        batcher3 = self.addHost('b3')
        filter1= self.addHost('f1')
        queue1 = self.addHost('q1')
        queue2 = self.addHost('q2')
        queue3 = self.addHost('q3')
        maintainer1 = self.addHost('m1')
        switch1 = self.addSwitch('s1')
        client = self.addHost('client')

        # Add links
        self.addLink(controller, switch1)
        self.addLink(app, switch1)
        self.addLink(batcher1, switch1)
        self.addLink(batcher2, switch1)
        self.addLink(batcher3, switch1)
        self.addLink(filter1, switch1)
        self.addLink(queue1, switch1)
        self.addLink(queue2, switch1)
        self.addLink(queue3, switch1)
        self.addLink(maintainer1, switch1)
        self.addLink(client, switch1)

def init(self, args):
    "init starts a single sinple cluster"
    args = args.split()
    if len(args) == 0:
        print('Please specify the topology')
        return
    topo = args[0]
    if len(args) < 2:
        gopath = os.path.join(os.getenv('HOME'), 'go')
    else:
        gopath = args[1]
    if topo == 'single':
        net = self.mn
        net.get('c').cmd(os.path.join(gopath, 'bin', 'gochariots-controller') + ' 8081 > c.log &')
        net.get('a').cmd(os.path.join(gopath, 'bin', 'gochariots-app') + ' 8080 > a.log &')
        net.get('b1').cmd(os.path.join(gopath, 'bin', 'gochariots-batcher') + ' 9000 > b.log &')
        net.get('f1').cmd(os.path.join(gopath, 'bin', 'gochariots-filter') + ' 9010 > f.log &')
        net.get('q1').cmd(os.path.join(gopath, 'bin', 'gochariots-queue') + ' 9020 true > q.log &')
        net.get('m1').cmd(os.path.join(gopath, 'bin', 'gochariots-maintainer') + ' 9030 > m.log &')

        c_ip = net.get('c').IP()
        a_ip = net.get('a').IP()
        b1_ip = net.get('b1').IP()
        f1_ip = net.get('f1').IP()
        q1_ip = net.get('q1').IP()
        m1_ip = net.get('m1').IP()
        print(net.get('client').cmd('curl -XPOST ' + a_ip + ':8080/batcher?host=' + b1_ip + ':9000'))
        print(net.get('client').cmd('curl -XPOST ' + c_ip + ':8081/batcher?host=' + b1_ip + ':9000'))
        print(net.get('client').cmd('curl -XPOST ' + c_ip + ':8081/filter?host=' + f1_ip + ':9010'))
        print(net.get('client').cmd('curl -XPOST ' + c_ip + ':8081/queue?host=' + q1_ip + ':9020'))
        print(net.get('client').cmd('curl -XPOST ' + c_ip + ':8081/maintainer?host=' + m1_ip + ':9030'))
    if topo == 'simple':
        net = self.mn
        net.get('c').cmd(os.path.join(gopath, 'bin', 'gochariots-controller') + ' 8081 > c.log &')
        net.get('a').cmd(os.path.join(gopath, 'bin', 'gochariots-app') + ' 8080 > a.log &')
        net.get('b1').cmd(os.path.join(gopath, 'bin', 'gochariots-batcher') + ' 9000 > b1.log &')
        net.get('b2').cmd(os.path.join(gopath, 'bin', 'gochariots-batcher') + ' 9001 > b2.log &')
        net.get('b3').cmd(os.path.join(gopath, 'bin', 'gochariots-batcher') + ' 9002 > b3.log &')
        net.get('f1').cmd(os.path.join(gopath, 'bin', 'gochariots-filter') + ' 9010 > f1.log &')
        net.get('q1').cmd(os.path.join(gopath, 'bin', 'gochariots-queue') + ' 9020 true > q1.log &')
        net.get('q2').cmd(os.path.join(gopath, 'bin', 'gochariots-queue') + ' 9021 false > q2.log &')
        net.get('q3').cmd(os.path.join(gopath, 'bin', 'gochariots-queue') + ' 9022 false > q3.log &')
        net.get('m1').cmd(os.path.join(gopath, 'bin', 'gochariots-maintainer') + ' 9030 > m1.log &')

        c_ip = net.get('c').IP()
        a_ip = net.get('a').IP()
        b1_ip = net.get('b1').IP()
        b2_ip = net.get('b2').IP()
        b3_ip = net.get('b3').IP()
        f1_ip = net.get('f1').IP()
        q1_ip = net.get('q1').IP()
        q2_ip = net.get('q2').IP()
        q3_ip = net.get('q3').IP()
        m1_ip = net.get('m1').IP()
        print(net.get('client').cmd('curl -XPOST ' + a_ip + ':8080/batcher?host=' + b1_ip + ':9000'))
        print(net.get('client').cmd('curl -XPOST ' + a_ip + ':8080/batcher?host=' + b2_ip + ':9001'))
        print(net.get('client').cmd('curl -XPOST ' + a_ip + ':8080/batcher?host=' + b3_ip + ':9002'))
        print(net.get('client').cmd('curl -XPOST ' + c_ip + ':8081/batcher?host=' + b1_ip + ':9000'))
        print(net.get('client').cmd('curl -XPOST ' + c_ip + ':8081/batcher?host=' + b2_ip + ':9001'))
        print(net.get('client').cmd('curl -XPOST ' + c_ip + ':8081/batcher?host=' + b3_ip + ':9002'))
        print(net.get('client').cmd('curl -XPOST ' + c_ip + ':8081/filter?host=' + f1_ip + ':9010'))
        print(net.get('client').cmd('curl -XPOST ' + c_ip + ':8081/queue?host=' + q1_ip + ':9020'))
        print(net.get('client').cmd('curl -XPOST ' + c_ip + ':8081/queue?host=' + q2_ip + ':9021'))
        print(net.get('client').cmd('curl -XPOST ' + c_ip + ':8081/queue?host=' + q3_ip + ':9022'))
        print(net.get('client').cmd('curl -XPOST ' + c_ip + ':8081/maintainer?host=' + m1_ip + ':9030'))

def log(self, node):
    "log shows the log of a node"
    net = self.mn
    try:
        n = net.get(node)
    except KeyError:
        print('Node ' + node + ' doesn\'t exist')
        return
    print(n.cmd('cat ' + node + '.log'))

def post(self, jsonfile):
    "post posts json file to the app"
    if jsonfile == '':
        print('Please specify json filename')
        return
    net = self.mn
    n = net.get('a')
    client = net.get('client')
    print(client.cmd('curl -i -XPOST -H "Content-Type: application/json" -d "@' + jsonfile + '" http://' + n.IP() + ':8080/record'))

def get(self, lid):
    "get gets the content of record"
    if lid == '':
        print('Please specify LId')
        return
    net = self.mn
    n = net.get('a')
    client = net.get('client')
    print(client.cmd('curl -XGET -H "Accept: application/json" http://' + n.IP() + ':8080/record/' + lid))

CLI.do_init = init
CLI.do_log = log
CLI.do_post = post
CLI.do_get = get

topos = {'single': (lambda: Single()), 'simple': (lambda: Simple())}