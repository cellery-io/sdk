# ------------------------------------------------------------------------
#
# Copyright 2019 WSO2, Inc. (http://wso2.com)
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
# http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License
#
# ------------------------------------------------------------------------

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
RUN dpkg -i ${USER_HOME}/debdir/ballerina-linux-installer-x64-1.0.3.deb
RUN dpkg -i ${USER_HOME}/debdir/cellery-ubuntu-x64-*.deb
# install required packages
RUN apt-get update && \
    DEBIAN_FRONTEND=noninteractive apt-get install -y --no-install-recommends \
curl \
gpg-agent \
software-properties-common

RUN curl -s https://packages.cloud.google.com/apt/doc/apt-key.gpg | apt-key add -
RUN add-apt-repository "deb [arch=amd64] http://apt.kubernetes.io/ kubernetes-xenial main"

RUN apt-get install kubectl

CMD ["/bin/bash"]
