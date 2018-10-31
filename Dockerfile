# Copyright 2017 Intel Corporation.
# Use of this source code is governed by a BSD-style
# license that can be found in the LICENSE file.

ARG USER_NAME
FROM ${USER_NAME}/nff-go-base

LABEL RUN docker run -it --privileged -v /sys/bus/pci/drivers:/sys/bus/pci/drivers -v /sys/kernel/mm/hugepages:/sys/kernel/mm/hugepages -v /sys/devices/system/node:/sys/devices/system/node -v /dev:/dev --name NAME -e NAME=NAME -e IMAGE=IMAGE IMAGE

RUN dnf -y install procps-ng iputils httpd wget; dnf clean all
#RUN sysctl -w net.ipv6.conf.all.disable_ipv6=0
RUN dd if=/dev/zero of=/var/www/html/10k.bin bs=1 count=10240
RUN dd if=/dev/zero of=/var/www/html/100k.bin bs=1 count=102400
RUN dd if=/dev/zero of=/var/www/html/1m.bin bs=1 count=1048576

WORKDIR /workdir
COPY nff-go-nat .
COPY client/client .
COPY config.json .
COPY config-vlan.json .
COPY config-dhcp.json .
COPY config-kni.json .
COPY config-kni-dhcp.json .
COPY config2ports.json .
