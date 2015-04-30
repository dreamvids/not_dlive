#!/bin/bash

FOLDER=/usr/local/dlive
BINARY=/usr/local/dlive/bin/dlive
USER=www-data

case "$1" in
	start)
		if ! [ -f "$FOLDER/logs/pid.txt" ]
		then
			su $USER -c "$BINARY --config $FOLDER/conf/config.json --log $FOLDER/logs/dlive.txt & echo \$! > /tmp/su.dlive.$$"
			cat /tmp/su.dlive.$$ > $FOLDER/logs/pid.txt
		fi
		;;
	stop)
		if [ -f "$FOLDER/logs/pid.txt" ]
		then
			kill -9 `cat "$FOLDER/logs/pid.txt"`
			rm "$FOLDER/logs/pid.txt"
		fi
		;;
	restart)
		if [ -f "$FOLDER/logs/pid.txt" ]
		then
			kill -9 `cat "$FOLDER/logs/pid.txt"`
			su $USER -c "$BINARY --config $FOLDER/conf/config.json --log $FOLDER/logs/dlive.txt & echo \$! > /tmp/su.dlive.$$"
			cat /tmp/su.dlive.$$ > $FOLDER/logs/pid.txt
		fi
		;;
esac
