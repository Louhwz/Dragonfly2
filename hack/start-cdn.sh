#!/usr/bin/env sh

set -o nounset
set -o errexit

nginx

/opt/dragonfly/df-cdn/cdn "$@"
