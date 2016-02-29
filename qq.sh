#!/bin/sh
qqlog="qq.log"
QQ="$TMPDIR$qqlog"
if [ $TMPDIR = "" ]; then
	if [ -f "/system/bin/adb" ]; then
		# android
		QQ="/data/local/tmp/$qqlog"
	else
		QQ="/tmp/$qqlog"
	fi
fi

if [ ! -f $QQ ]; then
	touch $QQ
fi

tail -100f $QQ
