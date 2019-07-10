FROM ubuntu:18.04

# set user configurations
ENV USER=cellery
ENV USER_ID=1000
ENV USER_GROUP=cellery
ENV USER_GROUP_ID=1000
ENV USER_HOME=/home/${USER}

# create a user group and a user
RUN groupadd --system -g ${USER_GROUP_ID} ${USER_GROUP} && \
    useradd --system --create-home --home-dir ${USER_HOME} --no-log-init -g ${USER_GROUP_ID} -u ${USER_ID} ${USER}

COPY files/*.deb ${USER_HOME}/debdir/
RUN dpkg -i ${USER_HOME}/debdir/ballerina-linux-installer-x64-0.991.0.deb
RUN dpkg -i ${USER_HOME}/debdir/cellery-ubuntu-x64.deb
# install required packages
RUN apt-get update && \
    DEBIAN_FRONTEND=noninteractive apt-get install -y --no-install-recommends \
curl \
gpg-agent \
software-properties-common

RUN curl -s https://packages.cloud.google.com/apt/doc/apt-key.gpg | apt-key add -
RUN add-apt-repository "deb [arch=amd64] http://apt.kubernetes.io/ kubernetes-xenial main"

RUN apt-get install kubectl
COPY files/cellery-*.jar /usr/lib/ballerina/ballerina-0.991.0/bre/lib/

CMD ["/bin/bash"]
