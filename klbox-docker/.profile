#!/bin/bash
# shellcheck source=/dev/null

# ~/.profile: executed by the command interpreter for login shells.
# This file is not read by bash(1), if ~/.bash_profile or ~/.bash_login
# exists.
# see /usr/share/doc/bash/examples/startup-files for examples.
# the files are located in the bash-doc package.

# the default umask is set in /etc/profile; for setting the umask
# for ssh logins, install and configure the libpam-umask package.
#umask 022

# source $HOME/.nix-profile/etc/profile.d/nix.sh
export PATH=$PATH:/home/kl/.nix-profile/bin

KL_LOCK_PATH=/home/kl/workspace/kl.lock
[ -f /kl-tmp/env ] && source /kl-tmp/env


if [ -n "$BASH_VERSION" ]; then
  # include .bashrc if it exists
  if [ -f "$HOME/.bashrc" ]; then
    . "$HOME/.bashrc"
  fi
fi

# set PATH so it includes user's private bin if it exists
if [ -d "$HOME/bin" ]; then
  PATH="$HOME/bin:$PATH"
fi

# set PATH so it includes user's private bin if it exists
if [ -d "$HOME/.local/bin" ]; then
  PATH="$HOME/.local/bin:$PATH"
fi

if [ -f "/kl-tmp/global-profile" ]; then
  source /kl-tmp/global-profile
fi

cd /home/kl/workspace || return
