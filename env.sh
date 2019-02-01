#!/bin/bash

if [ -z "${NFF_GO}" ]
then
	echo "You need to define NFF_GO variable which points to root of built NFF_GO repository."
	exit 1
fi

export RTE_TARGET=x86_64-native-linuxapp-gcc

DPDK_DIR=dpdk
DPDK_INSTALL_DIR=${RTE_TARGET}-install
export RTE_SDK="${NFF_GO}"/dpdk/${DPDK_DIR}/${DPDK_INSTALL_DIR}/usr/local/share/dpdk

export CGO_LDFLAGS_ALLOW='-Wl,--((no-)?whole-archive|((start|end)-group))'
export CGO_CFLAGS="-I${RTE_SDK}/${RTE_TARGET}/include -O3 -std=gnu11 -m64 -pthread -march=native -mno-fsgsbase -mno-f16c -DRTE_MACHINE_CPUFLAG_SSE -DRTE_MACHINE_CPUFLAG_SSE2 -DRTE_MACHINE_CPUFLAG_SSE3 -DRTE_MACHINE_CPUFLAG_SSSE3 -DRTE_MACHINE_CPUFLAG_SSE4_1 -DRTE_MACHINE_CPUFLAG_SSE4_2 -DRTE_MACHINE_CPUFLAG_PCLMULQDQ -DRTE_MACHINE_CPUFLAG_RDRAND -DRTE_MACHINE_CPUFLAG_F16C -include rte_config.h -Wno-deprecated-declarations"
export CGO_LDFLAGS="-L${RTE_SDK}/${RTE_TARGET}/lib -Wl,--no-as-needed -Wl,-export-dynamic"

if ! command -v protoc &> /dev/null; then
	echo You should install protobuf compiler package, e.g. \"sudo dnf install protobuf-compiler\" or \"sudo apt-get install protobuf-compiler\"
fi
if ! command -v protoc-gen-go &> /dev/null; then
	echo You should install Go plugin for protobuf compiler with \"go get github.com/golang/protobuf/protoc-gen-go\" and add target directory to PATH \(\$GOPATH/bin or \~/go/bin\)
fi

