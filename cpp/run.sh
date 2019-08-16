#!/bin/bash
clear
rm -r bin
mkdir bin

gcc convert_stream.c -o bin/convert \
-I/usr/local/ffmpeg/include \
-L/usr/local/ffmpeg/lib -lavcodec -lavformat -lavfilter -lavutil -lavdevice -lswresample

./bin/convert