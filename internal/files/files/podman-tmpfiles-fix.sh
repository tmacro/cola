#!/usr/bin/env bash

# This script is a workaround for the podman extension using systemd-tmfiles to copy the default config to /etc/containers
# When ignition is used to create container configs in /etc/containers/systemd then systemd-tmpfiles will not copy the default config and podman will fail

# Check if /etc/containers/policy.json exists
if [ ! -f /etc/containers/policy.json ]; then
    # Copy the default config
    cp -r  /usr/share/podman/etc/containers /etc/
fi
