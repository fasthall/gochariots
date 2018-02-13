def generate(num_dc, id, filename):
    with open(filename, 'w') as outfile:
        outfile.write('controller: "controller:8081"\n')
        outfile.write('num_dc: {}\n'.format(num_dc))
        outfile.write('id: {}\n'.format(id))