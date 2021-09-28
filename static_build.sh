#! /bin/sh
#
# static_build.sh
# Copyright (C) 2020 forseason <me@forseason.vip>
#
# Distributed under terms of the MIT license.
#


CGO_ENABLED=0 go build -a -ldflags '-extldflags "-static"'  .
