# Using DynamoDB as storage

Instead of using FLStore, user can opt for using AWS DynamoDB as storage.

Modify [compose_dynamodb.yaml](../deploy/compose_dynamodb.yaml). Replace `AWS_ACCESS_KEY_ID` and `AWS_SECRET_ACCESS_KEY` with your credential information. Run

    $ docker-compose -f compose_dynamodb.yaml up -d

to start GoChariots.