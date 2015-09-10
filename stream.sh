#!/bin/bash

#ffmpeg -f video4linux2 -s 640x480 -r 25 -i /dev/video0 -f webm http://127.0.0.1:8081/stream/push/mdr
ffmpeg -f video4linux2 -s 640x480 -r 25 -i /dev/video0 -f webm video.webm
