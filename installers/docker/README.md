#### How to buid the Cellery docker image.
```
docker build -t celleryio/cellery-cli:<version> .
```

#### How to run the Cellery CLI using docker.

celleryio/cellery-cli image is built with a user named cellery with uid 1000 and gid 1000.

Before run the cellery-cli docker image, user needs to map the current user id (UID) to the user id which used in the docker file.

If the user has a different user id other than 1000, user needs to rebuild the docker file with that uid.

```
docker run -it -u $(id -u):$(id -g) \
--mount type=bind,source=/home/<user>,\
target=/home/cellery  \
celleryio/cellery-cli
```
