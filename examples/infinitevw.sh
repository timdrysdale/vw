#!/bin/sh 
pkill -9 vw
for (( ; ; ))
do
../vw stream
pkill -9 vw
done
