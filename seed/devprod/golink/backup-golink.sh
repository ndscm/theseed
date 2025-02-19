#/usr/bin/env bash
# Copy this script to /etc/cron.daily
set -eux
set -o pipefail

mkdir -p /mnt/data/golink/backup
datetime=$(date +%Y%m%d%H%M%S)
cp /mnt/data/golink/golink.db /mnt/data/golink/backup/golink-${datetime}.db
