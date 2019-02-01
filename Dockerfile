# Copyright 2017-2018 Intel Corporation.
# Use of this source code is governed by a BSD-style
# license that can be found in the LICENSE file.

ARG USER_NAME
FROM ${USER_NAME}/nff-go-base

LABEL RUN docker run -it --privileged -v /sys/bus/pci/drivers:/sys/bus/pci/drivers -v /sys/kernel/mm/hugepages:/sys/kernel/mm/hugepages -v /sys/devices/system/node:/sys/devices/system/node -v /dev:/dev --name NAME -e NAME=NAME -e IMAGE=IMAGE IMAGE

RUN apt-get install -y procps iproute2 iputils-ping net-tools apache2 wget; apt-get clean
RUN dd if=/dev/zero of=/var/www/html/10k.bin bs=1 count=10240
RUN dd if=/dev/zero of=/var/www/html/100k.bin bs=1 count=102400
RUN dd if=/dev/zero of=/var/www/html/1m.bin bs=1 count=1048576

# The following command cannot be executed at build stage but is
# required for IPv6 KNI interfaces, so it should be executed when
# container is ran
# RUN sysctl -w net.ipv6.conf.all.disable_ipv6=0

WORKDIR /workdir

# NAT executables
COPY nff-go-nat .
COPY client/client .

# Test applications
COPY test/httpperfserv/httpperfserv .
COPY test/wrk/wrk .

# Configs without VLANs
COPY config.json .
COPY config-vlan.json .
COPY config-dhcp.json .

# Configs with VLANs
COPY config-kni.json .
COPY config-kni-vlan.json .
COPY config-kni-dhcp.json .

# Two ports config
COPY config2ports.json .
