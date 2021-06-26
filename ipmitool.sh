#!/bin/sh
echo $@ >> log

/usr/bin/ipmitool $@
