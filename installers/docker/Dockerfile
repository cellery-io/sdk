FROM ubuntu:18.04

# set user configurations
ARG USER=cellery
ARG USER_ID=802
ARG USER_GROUP=cellery
ARG USER_GROUP_ID=802
ARG USER_HOME=/home/${USER}
# set dependant files directory
ARG FILES=./files

# create a user group and a user
RUN groupadd --system -g ${USER_GROUP_ID} ${USER_GROUP} && \
    useradd --system --create-home --home-dir ${USER_HOME} --no-log-init -g ${USER_GROUP_ID} -u ${USER_ID} ${USER}

COPY install.sh /${USER_HOME}/
COPY files/*.deb ${USER_HOME}/debdir/

RUN mkdir -p /${USER_HOME}/workspace
 
RUN chmod +x /${USER_HOME}/install.sh

RUN /${USER_HOME}/install.sh

CMD ["/bin/bash"]
