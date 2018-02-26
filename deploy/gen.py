import stack_gen
import config_gen

stack_gen.mongodb_host = ''
stack_gen.generate('stack.yaml')

num_lowgo = 1
for i in range(num_lowgo):
    filename = 'config{}.yaml'.format(i + 1)
    config_gen.generate(num_lowgo, i + 1, filename)
