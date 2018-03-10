#!/bin/bash

FILE="/home/public/forts-and-ruins-online/server.go"

REPDIR="/home/public/forts-and-ruins-online/"

SUM=$(cksum $FILE)
go run $FILE &
while true; do

	# check if server file has changed. 
	# If it has, hard restart it
	TEMPSUM=$(cksum $FILE)
	if [ "$SUM" != "$TEMPSUM" ];
	then
		echo "RESTARTING"
		pkill \(go\)\|\(server\)
		wait $!
		go run $FILE &
	fi
	SUM=$TEMPSUM

	# Try to git pull. If these start failing,
	# I'll pipe "yes" into it
	git -C $REPDIR pull

	# don't need to bother github too much, do we? ;)
	# vim protip: ctr-a increments, ctr-x decrements. Works on hex and bin, too
	sleep 5
done

trap "killall go" SIGINT SIGTERM
