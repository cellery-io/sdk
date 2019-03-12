###How to run the Cellery CLI using docker.

```
docker run -it -u -u $(id -u):$(id -g) \
--mount type=bind,source=/home/<user>,\
target=/home/cellery  \
celleryio/cellery-cli
```
