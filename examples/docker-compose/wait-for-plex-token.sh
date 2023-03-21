#!/usr/bin/env sh

for i in {1..60}; do
  PLEX_TOKEN=`xmlstarlet sel -T -t -m "/Preferences" -v "@PlexOnlineToken" -n "/config/Library/Application Support/Plex Media Server/Preferences.xml"`
  if [ -z $PLEX_TOKEN ]
  then
    sleep 5
  else
    echo "export PLEX_TOKEN=$PLEX_TOKEN" > /plex/token && break
  fi
done

if [ -z $PLEX_TOKEN ]; then exit 1; fi
