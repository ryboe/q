#!/usr/bin/env bash
set -euo pipefail

logfile="q"
logpath=$TMPDIR/$logfile

if [[ -z "$TMPDIR" ]]; then
	if [[ -e "/system/bin/adb" ]]; then
		# android
		logpath="/data/local/tmp/$logfile"
	else
		logpath="/tmp/$logfile"
	fi
fi

if [[ ! -f "$logpath" ]]; then
	touch $logpath
fi

tail -100f $logpath
