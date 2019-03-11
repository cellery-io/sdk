###How to run the Cellery CLI using docker.

```
docker run -it -u cellery \
--mount type=bind,source=/home/<user>/works/cellery,\
target=/home/cellery/workspace  \
celleryio/cellery-cli
```
