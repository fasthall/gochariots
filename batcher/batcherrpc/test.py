import grpc

import batcherrpc_pb2
import batcherrpc_pb2_grpc

def run():
    channel = grpc.insecure_channel('localhost:9000')
    stub = batcherrpc_pb2_grpc.BatcherRPCStub(channel)
    response = stub.ReceiveRecord(batcherrpc_pb2.RPCRecord(id='', host=0, trace='123', seqid=0, depth=0))
    print("Greeter client received: " + response.message)

if __name__ == '__main__':
    run()