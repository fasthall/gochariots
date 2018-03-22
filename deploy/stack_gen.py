import yaml
from collections import OrderedDict

ports = {
    'app': '8080',
    'controller': '8081',
    'batcher': '9000',
    'batcher_lb': '9000',
    'cache': '6380',
    'storage': '27017',
}

mongodb_host = ''

def ordered_load(stream, Loader=yaml.Loader, object_pairs_hook=OrderedDict):
    class OrderedLoader(Loader):
        pass
    def construct_mapping(loader, node):
        loader.flatten_mapping(node)
        return object_pairs_hook(loader.construct_pairs(node))
    OrderedLoader.add_constructor(
        yaml.resolver.BaseResolver.DEFAULT_MAPPING_TAG,
        construct_mapping)
    return yaml.load(stream, OrderedLoader)

def ordered_dump(data, stream=None, Dumper=yaml.Dumper, **kwds):
    class OrderedDumper(Dumper):
        pass
    def _dict_representer(dumper, data):
        return dumper.represent_mapping(
            yaml.resolver.BaseResolver.DEFAULT_MAPPING_TAG,
            data.items())
    OrderedDumper.add_representer(OrderedDict, _dict_representer)
    return yaml.dump(data, stream, OrderedDumper, **kwds)

def replace_port(stack_file, service):
    port_name = service + '_port'
    if 'entrypoint' in stack_file['services'][service] and service in ports:
        stack_file['services'][service]['entrypoint'] = stack_file['services'][service]['entrypoint'].replace(port_name, ports[service])
    if 'ports' in stack_file['services'][service] and service in ports:
        for i in range(len(stack_file['services'][service]['ports'])):
            if ':' in stack_file['services'][service]['ports'][i]:
                stack_file['services'][service]['ports'][i] = stack_file['services'][service]['ports'][i].replace(port_name, ports[service])
            else:
                stack_file['services'][service]['ports'][i] = int(ports[service])
    if 'environment' in stack_file['services'][service]:
        for i in range(len(stack_file['services'][service]['environment'])):
            if service in ports:
                stack_file['services'][service]['environment'][i] = stack_file['services'][service]['environment'][i].replace(port_name, ports[service])
            stack_file['services'][service]['environment'][i] = stack_file['services'][service]['environment'][i].replace('mongodb_host', mongodb_host)

def generate(outfile_name):
    with open('stack_template.yaml', 'r') as infile:
        stack_file = ordered_load(infile, yaml.SafeLoader)
        replace_port(stack_file, 'app')
        replace_port(stack_file, 'controller')
        replace_port(stack_file, 'batcher')
        replace_port(stack_file, 'queue-leader')
        replace_port(stack_file, 'queue')
        replace_port(stack_file, 'maintainer')
        replace_port(stack_file, 'batcher_lb')
        replace_port(stack_file, 'cache')
        replace_port(stack_file, 'storage')

    with open(outfile_name, 'w') as outfile:
        ordered_dump(stack_file, outfile, default_flow_style=False)
