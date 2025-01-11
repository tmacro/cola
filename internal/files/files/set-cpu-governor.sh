#!/usr/bin/env bash

# Source the default configuration file
. /etc/default/cpu_governor

if [ -z "$POWER_PROFILE" ]; then
    echo "POWER_PROFILE is not set. Exiting..."
    exit 0
fi


# See if a we can even set the governor
govs=$(find /sys/devices/system/cpu/ -name "scaling_governor" | wc -l)
if [ "$govs" -eq 0 ]; then
    echo "No CPU scaling governor found. Exiting..."
    exit 0
fi

# Load the module
/usr/sbin/modprobe cpufreq_${POWER_PROFILE}

# Set the governor
echo "$POWER_PROFILE" | tee /sys/devices/system/cpu/cpu*/cpufreq/scaling_governor
