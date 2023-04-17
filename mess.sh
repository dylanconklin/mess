#!/bin/bash

if ! [ -z "$1" ] && ! [ -z "$2" ]
then
    telnet -E "$1" "$2"
fi
echo "USAGE: mess IP_ADDR PORT"
