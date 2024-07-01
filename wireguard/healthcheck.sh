#!/bin/bash

# Check if the target IP is reachable
if ping -c 1 100.64.0.1 > /dev/null 2>&1
then
    exit 0  # IP is reachable, healthcheck passes
else
    exit 1  # IP is not reachable, healthcheck fails
fi
