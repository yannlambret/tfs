#!/bin/sh
set -e

# Run tfs to fetch/activate proper Terraform binary
tfs

# Execute Terraform with passed arguments
exec "$@"
