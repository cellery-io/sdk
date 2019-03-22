#### How to run the Cellery CLI using docker.

User needs to map the user id to the user id use in the docker file.

celleryio/cellery-cli image is built with a user named cellery with uid 1000 and gid 1000.

If the user different user id other than 1000, user needs to rebuild the docker file with that uid.
```
docker run -it -u $(id -u):$(id -g) \
--mount type=bind,source=/home/<user>,\
target=/home/cellery  \
celleryio/cellery-cli
```
