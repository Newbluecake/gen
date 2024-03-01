#!/bin/bash

set -exuo pipefail

if [ -z "$1" ]; then
	exit
fi

export CODENAME=$(lsb_release --codename --short)
export LLVM_VERSION=$1

# 1. download llvm from https://github.com/llvm/llvm-project/releases
# 2. extract to /usr/lib/llvm-$LLVM_VERSION

sudo rm -f /usr/bin/clang
sudo rm -f /usr/bin/clang++
sudo rm -f /usr/bin/llvm-config
sudo ln -s /usr/lib/llvm-$LLVM_VERSION/bin/clang /usr/bin/clang
sudo ln -s /usr/lib/llvm-$LLVM_VERSION/bin/clang++ /usr/bin/clang++
sudo ln -s /usr/lib/llvm-$LLVM_VERSION/bin/llvm-config /usr/bin/llvm-config
sudo ldconfig
