#!/bin/sh
# Redirect apt at the offline mirror at /apt/mirror. apt_mirror already generated
# /apt/mirror/sources.list.d/debian.sources pointing at file:///apt/mirror; this
# drop-in makes apt read it (ignoring /etc/apt/sources.list.d) until restore.sh
# removes it.
set -eu

cat >/etc/apt/apt.conf.d/00mirror <<'CONF'
Dir::Etc::sourceparts "/apt/mirror/sources.list.d";
CONF
