#!/bin/sh
# Undo use.sh: stop redirecting apt at the mirror (it falls back to the untouched
# network debian.sources), rename the offline file:// lists to the names a real
# fetch would produce so the cache matches the http source, then drop the mirror.
set -eu

rm -f /etc/apt/apt.conf.d/00mirror

# apt names each list file after its source URI (file:///apt/mirror -> the
# `_apt_mirror_` prefix). Map each suite back to its original archive URI and
# rename.
awk '/^URIs:/{u=$2} /^Suites:/{for(i=2;i<=NF;i++)m[$i]=u} END{for(s in m)print s,m[s]}' \
  /etc/apt/sources.list.d/debian.sources |
  while read suite uri; do
    prefix=$(printf '%s' "$uri" | sed 's|^[a-z]\+://||; s|/|_|g')
    for f in /var/lib/apt/lists/_apt_mirror_dists_"$suite"_*; do
      [ -e "$f" ] || continue
      rest=$(basename "$f")
      rest=${rest#_apt_mirror_dists_${suite}_}
      mv "$f" "/var/lib/apt/lists/${prefix}_dists_${suite}_${rest}"
    done
  done
