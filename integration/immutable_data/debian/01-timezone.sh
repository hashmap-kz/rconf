#!/bin/bash
set -euo pipefail

default_timezone='Asia/Aqtau'
echo "${default_timezone}" >/etc/timezone
ln -sf "/usr/share/zoneinfo/${default_timezone}" /etc/localtime
hwclock --systohc
dpkg-reconfigure -f noninteractive tzdata
