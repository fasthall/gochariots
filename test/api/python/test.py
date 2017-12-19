import grpc
import batcherrpc_pb2
import batcherrpc_pb2_grpc

channel = grpc.insecure_channel('localhost:9000')
stub = batcherrpc_pb2_grpc.BatcherRPCStub(channel)

records = batcherrpc_pb2.RPCRecords(
    records = [
        batcherrpc_pb2.RPCRecord(
            host = 1,
            tag = {'k1', 'v1'}
        ),
        batcherrpc_pb2.RPCRecord(
            host = 1,
            tag = {'k2', 'v2'}
        )
    ]
)

stub.TOIDReceiveRecords(records)

