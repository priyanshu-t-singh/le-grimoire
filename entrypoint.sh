#!/bin/sh
set -e

# Fix ownership of the mounted config dir every startup
chown -R appuser:appgroup /home/appuser/.config

# Use passed args if provided, otherwise default to "serve"
if [ "$#" -eq 0 ]; then
    set -- serve
fi

exec su-exec appuser:appgroup le-grimoire "$@"
