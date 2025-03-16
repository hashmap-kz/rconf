#!/bin/bash
set -euo pipefail

default_locale='en_US.UTF-8'
sed -i -e 's/# en_US.UTF-8 UTF-8/en_US.UTF-8 UTF-8/' /etc/locale.gen
sed -i -e 's/# ru_RU.UTF-8 UTF-8/ru_RU.UTF-8 UTF-8/' /etc/locale.gen
echo "LANG=${default_locale}" >/etc/default/locale
dpkg-reconfigure --frontend=noninteractive locales
update-locale LANG="${default_locale}"
