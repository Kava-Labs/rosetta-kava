#!/usr/bin/env bash

export START_KAVA_NODE=true

if [ "$MODE" = "offline" ]
then
  START_KAVA_NODE=false
fi

supervisord -c /etc/supervisor/supervisord.conf
