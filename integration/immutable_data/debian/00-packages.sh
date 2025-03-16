#!/bin/bash
set -euo pipefail

apt update -y
apt install -y \
  chrony \
  cron \
  mc \
  ncdu \
  net-tools \
  openssh-server \
  rsyslog \
  sudo \
  tar \
  unzip \
  util-linux-extra \
  vim
