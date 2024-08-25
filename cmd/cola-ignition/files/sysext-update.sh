#!/usr/bin/env bash

set -e -o pipefail

DATA_DIR="/opt/sysext-update"

function gather_sysext_versions() {
    echo "Gathering sysext versions into $1"
    find  /etc/extensions/ -type l -exec realpath {} \; | grep ^/opt/extensions/ | xargs basename --multiple | sort > "$1"
}

function compare_sysext_versions() {
    diff -q "$1" "$2" > /dev/null
    return $?
}

if [ ! -d "$DATA_DIR" ]; then
    mkdir -p "$DATA_DIR"
fi

CMD="$1"

if [ -z "$CMD" ]; then
    echo "Usage: $0 <pre-update|post-update>"
    exit 1
fi

case "$CMD" in
    "pre-update")
        gather_sysext_versions "$DATA_DIR/versions.before"
        ;;
    "post-update")
        if [ ! -f "$DATA_DIR/versions.before" ]; then
            echo "No previous sysext versions found. Skipping check."
            exit 0
        fi
        gather_sysext_versions "$DATA_DIR/versions.after"
        if compare_sysext_versions "$DATA_DIR/versions.before" "$DATA_DIR/versions.after"; then
            echo "No changes in sysext versions"
        else
            echo "Changes in sysext versions detected. Scheduling a restart."
            /usr/bin/locksmithctl send-need-reboot
        fi
        ;;
    *)
        echo "Unknown command: $CMD"
        exit 1
        ;;
esac
