#!/bin/sh
# Wrapper for install.sh which provides kolvillag-specific configurations.

SBP_LINUX_CONFIG=~/sbp-linux-config
SRC=$SBP_LINUX_CONFIG/hosts/kolvillag/src
$SBP_LINUX_CONFIG/install.sh $SRC