# Using Google Datastore as storage

Instead of using FLStore, user can opt for using Google Datastore as storage.

Copy your Google Cloud Platform service credential json file to `cred/cred.json`.

Modify [compose_datastore.yaml](../deploy/compose_datastore.yaml). Replace `DATASTORE_PROJECT_ID` with your GCP project ID. Run

    $ docker-compose -f compose_datastore.yaml up -d

to start GoChariots.