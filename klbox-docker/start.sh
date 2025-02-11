#!/usr/bin/env bash
# shellcheck source=/dev/null
set -o errexit
set -o pipefail

trap "echo kloudlite-entrypoint:CRASHED >&2" EXIT SIGINT SIGTERM

#/nix-installer/install

export IN_DEV_BOX="true"
export KL_WORKSPACE="$KL_WORKSPACE"
export KL_TMP_PATH="/kl-tmp"

cat <<EOL >/kl-tmp/global-profile
export SSH_PORT=$SSH_PORT
export IN_DEV_BOX="true"
export MAIN_PATH=$PATH
export KL_TMP_PATH="/kl-tmp"
export PLATFORM_ARCH=$(uname -m)

export KL_WORKSPACE="$KL_WORKSPACE"
export KL_BASE_URL="$KL_BASE_URL"
# export KLCONFIG_PATH="$KLCONFIG_PATH"
export SHELL=$(which zsh)
export KL_BOX_MODE="true"
EOL

chown -R kl /kl-tmp/global-profile

entrypoint_executed="/home/kl/.kloudlite_entrypoint_executed"
if [ ! -f "$entrypoint_executed" ]; then
  mkdir -p /home/kl/.config
  cp /tmp/.zshrc /home/kl/.zshrc
  cp /tmp/.bashrc /home/kl/.bashrc
  cp /tmp/.profile /home/kl/.profile
  cp /tmp/.check-online /home/kl/.check-online
  ln -sf /home/kl/.profile /home/kl/.zprofile
  cp /tmp/aliasrc /home/kl/.config/aliasrc
  echo "successfully initialized .profile and .bashrc" >>$entrypoint_executed
  # ssh-keygen -t rsa -b 4096 -f ~/.ssh/id_rsa -N "" <<<y >/dev/null 2>&1
fi

export PATH=$PATH:/home/kl/.nix-profile/bin

trap - EXIT SIGTERM SIGINT
echo "kloudlite-entrypoint:SETUP_COMPLETE"
