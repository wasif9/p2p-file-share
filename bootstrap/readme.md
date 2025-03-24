# Bootstrap Node
A starting node that is used for the first connection of a new node.

## Run with Docker
1. ``cd ..`` to go to the root directory of the project since we need to copy ``go.mod`` and ``go.sum``.
2. ``docker build -t bootstrap -f .\bootstrap\Dockerfile .`` to build the image.
3. ``docker run -p 4001:4001 --name bootstrap bootstrap`` to build and run the container.
4. Crate an ``.env`` environmnet configuration file under the ``peer-app`` directory.
5. Add the following contents to the ``.env`` file\
``BOOTSTRAP_ADDR=<Bootstrap multiaddrs>``\
where ``<Bootstrap multiaddrs>`` is the address of the bootstrap node. (.e.,g ``/ip4/127.0.0.1/tcp/4001/p2p/<Node ID>``)


## Run with Google Cloud Platform (Static IP)
1. Crate an ``.env`` environmnet configuration file under the ``peer-app`` directory.
5. Add the following contents to the ``.env`` file\
``BOOTSTRAP_ADDR=/ip4/<Static IP>/tcp/4001/<Node ID>``

## Node ID
The bootstrap's node ID is depended on the ``base64PrivKey``.