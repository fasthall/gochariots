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
        filter = self.addHost('f1')
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
        filter1 = self.addHost('f1')
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

class Double(Topo):
    def __init__(self):
        Topo.__init__(self)

        # Cluster 0
        app01 = self.addHost('a01')
        controller01 = self.addHost('c01')
        batcher01 = self.addHost('b01')
        filter01 = self.addHost('f01')
        filter02 = self.addHost('f02')
        queue01 = self.addHost('q01')
        maintainer01 = self.addHost('m01')
        switch0 = self.addSwitch('s0')
        client0 = self.addHost('client0')
        self.addLink(app01, switch0)
        self.addLink(controller01, switch0)
        self.addLink(batcher01, switch0)
        self.addLink(filter01, switch0)
        self.addLink(filter02, switch0)
        self.addLink(queue01, switch0)
        self.addLink(maintainer01, switch0)
        self.addLink(client0, switch0)

        # Cluster 1
        app11 = self.addHost('a11')
        controller11 = self.addHost('c11')
        batcher11 = self.addHost('b11')
        filter11 = self.addHost('f11')
        filter12 = self.addHost('f12')
        queue11 = self.addHost('q11')
        maintainer11 = self.addHost('m11')
        switch1 = self.addSwitch('s1')
        client1 = self.addHost('client1')
        self.addLink(app11, switch1)
        self.addLink(controller11, switch1)
        self.addLink(batcher11, switch1)
        self.addLink(filter11, switch1)
        self.addLink(filter12, switch1)
        self.addLink(queue11, switch1)
        self.addLink(maintainer11, switch1)
        self.addLink(client1, switch1)

        # Connect the switches
        switch3 = self.addSwitch('s3')
        self.addLink(switch0, switch3)
        self.addLink(switch1, switch3)

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
    net = self.mn
    if topo == 'single':
        net.get('c').cmd(os.path.join(gopath, 'bin', 'gochariots-controller') + ' 8081 1 0 > c.log &')
        net.get('a').cmd(os.path.join(gopath, 'bin', 'gochariots-app') + ' 8080 1 0 > a.log &')
        net.get('b1').cmd(os.path.join(gopath, 'bin', 'gochariots-batcher') + ' 9000 1 0 > b.log &')
        net.get('f1').cmd(os.path.join(gopath, 'bin', 'gochariots-filter') + ' 9010 1 0 > f.log &')
        net.get('q1').cmd(os.path.join(gopath, 'bin', 'gochariots-queue') + ' 9020 1 0 true > q.log &')
        net.get('m1').cmd(os.path.join(gopath, 'bin', 'gochariots-maintainer') + ' 9030 1 0 > m.log &')

        # Wait for controller and app running
        time.sleep(2)

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
        net.get('c').cmd(os.path.join(gopath, 'bin', 'gochariots-controller') + ' 8081 1 0 > c.log &')
        net.get('a').cmd(os.path.join(gopath, 'bin', 'gochariots-app') + ' 8080 1 0 > a.log &')
        net.get('b1').cmd(os.path.join(gopath, 'bin', 'gochariots-batcher') + ' 9000 1 0 > b1.log &')
        net.get('b2').cmd(os.path.join(gopath, 'bin', 'gochariots-batcher') + ' 9001 1 0 > b2.log &')
        net.get('b3').cmd(os.path.join(gopath, 'bin', 'gochariots-batcher') + ' 9002 1 0 > b3.log &')
        net.get('f1').cmd(os.path.join(gopath, 'bin', 'gochariots-filter') + ' 9010 1 0 > f1.log &')
        net.get('q1').cmd(os.path.join(gopath, 'bin', 'gochariots-queue') + ' 9020 1 0 true > q1.log &')
        net.get('q2').cmd(os.path.join(gopath, 'bin', 'gochariots-queue') + ' 9021 1 0 false > q2.log &')
        net.get('q3').cmd(os.path.join(gopath, 'bin', 'gochariots-queue') + ' 9022 1 0 false > q3.log &')
        net.get('m1').cmd(os.path.join(gopath, 'bin', 'gochariots-maintainer') + ' 9030 1 0 > m1.log &')

        # Wait for controller and app running
        time.sleep(2)

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
    if topo == 'double':
        net.get('a01').cmd(os.path.join(gopath, 'bin', 'gochariots-app') + ' 8080 2 0 > a01.log &')
        net.get('c01').cmd(os.path.join(gopath, 'bin', 'gochariots-controller') + ' 8081 2 0 > c01.log &')
        net.get('b01').cmd(os.path.join(gopath, 'bin', 'gochariots-batcher') + ' 9000 2 0 > b01.log &')
        net.get('f01').cmd(os.path.join(gopath, 'bin', 'gochariots-filter') + ' 9010 2 0 > f01.log &')
        net.get('f02').cmd(os.path.join(gopath, 'bin', 'gochariots-filter') + ' 9011 2 0 > f02.log &')
        net.get('q01').cmd(os.path.join(gopath, 'bin', 'gochariots-queue') + ' 9020 2 0 true > q01.log &')
        net.get('m01').cmd(os.path.join(gopath, 'bin', 'gochariots-maintainer') + ' 9030 2 0 > m01.log &')
        net.get('a11').cmd(os.path.join(gopath, 'bin', 'gochariots-app') + ' 8080 2 1 > a11.log &')
        net.get('c11').cmd(os.path.join(gopath, 'bin', 'gochariots-controller') + ' 8081 2 1 > c11.log &')
        net.get('b11').cmd(os.path.join(gopath, 'bin', 'gochariots-batcher') + ' 9100 2 1 > b11.log &')
        net.get('f11').cmd(os.path.join(gopath, 'bin', 'gochariots-filter') + ' 9110 2 1 > f11.log &')
        net.get('f12').cmd(os.path.join(gopath, 'bin', 'gochariots-filter') + ' 9111 2 1 > f12.log &')
        net.get('q11').cmd(os.path.join(gopath, 'bin', 'gochariots-queue') + ' 9120 2 1 true > q11.log &')
        net.get('m11').cmd(os.path.join(gopath, 'bin', 'gochariots-maintainer') + ' 9130 2 1 > m11.log &')

        # Wait for controller and app running
        time.sleep(2)

        a01_ip = net.get('a01').IP()
        c01_ip = net.get('c01').IP()
        b01_ip = net.get('b01').IP()
        f01_ip = net.get('f01').IP()
        f02_ip = net.get('f02').IP()
        q01_ip = net.get('q01').IP()
        m01_ip = net.get('m01').IP()
        a11_ip = net.get('a11').IP()
        c11_ip = net.get('c11').IP()
        b11_ip = net.get('b11').IP()
        f11_ip = net.get('f11').IP()
        f12_ip = net.get('f12').IP()
        q11_ip = net.get('q11').IP()
        m11_ip = net.get('m11').IP()

        print(net.get('client0').cmd('curl -XPOST ' + a01_ip + ':8080/batcher?host=' + b01_ip + ':9000'))
        print(net.get('client0').cmd('curl -XPOST ' + c01_ip + ':8081/batcher?host=' + b01_ip + ':9000'))
        print(net.get('client0').cmd('curl -XPOST ' + c01_ip + ':8081/filter?host=' + f01_ip + ':9010'))
        print(net.get('client0').cmd('curl -XPOST ' + c01_ip + ':8081/filter?host=' + f02_ip + ':9011'))
        print(net.get('client0').cmd('curl -XPOST ' + c01_ip + ':8081/queue?host=' + q01_ip + ':9020'))
        print(net.get('client0').cmd('curl -XPOST ' + c01_ip + ':8081/maintainer?host=' + m01_ip + ':9030'))

        print(net.get('client1').cmd('curl -XPOST ' + a11_ip + ':8080/batcher?host=' + b11_ip + ':9100'))
        print(net.get('client1').cmd('curl -XPOST ' + c11_ip + ':8081/batcher?host=' + b11_ip + ':9100'))
        print(net.get('client1').cmd('curl -XPOST ' + c11_ip + ':8081/filter?host=' + f11_ip + ':9110'))
        print(net.get('client1').cmd('curl -XPOST ' + c11_ip + ':8081/filter?host=' + f12_ip + ':9111'))
        print(net.get('client1').cmd('curl -XPOST ' + c11_ip + ':8081/queue?host=' + q11_ip + ':9120'))
        print(net.get('client1').cmd('curl -XPOST ' + c11_ip + ':8081/maintainer?host=' + m11_ip + ':9130'))

        print(net.get('client0').cmd('curl -XPOST ' + c01_ip + ':8081/remote/batcher?dc=1\&host=' + b11_ip + ':9100'))
        print(net.get('client1').cmd('curl -XPOST ' + c11_ip + ':8081/remote/batcher?dc=0\&host=' + b01_ip + ':9000'))

    print('If anything wrong, try to restart mininet and run init again with GOPATH specified.')
    print('Example: mininet> init single /home/vagrant/go')
def log(self, node):
    "log shows the log of a node"
    net = self.mn
    try:
        n = net.get(node)
    except KeyError:
        print('Node ' + node + ' doesn\'t exist')
        return
    print(n.cmd('cat ' + node + '.log'))

def post(self, args):
    "post posts json file to the app"
    args = args.split()
    if len(args) == 0:
        print('Please specify json filename')
        return
    jsonfile = args[0]
    net = self.mn
    if len(args) == 2:
        n = net.get('a' + args[1] + '1')
        client = net.get('client' + args[1])
    else:
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

topos = {
    'single': (lambda: Single()),
    'simple': (lambda: Simple()),
    'double': (lambda: Double())
    }