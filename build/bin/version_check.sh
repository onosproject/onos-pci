#!/usr/bin/env bash
# SPDX-License-Identifier: Apache-2.0
# Copyright 2024 Intel Corporation

# check if version format is matched to SemVer
VER_REGEX='^(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)$'
if [[ ! $(cat VERSION | tr -d '\n' | tr -d "\-dev") =~ $VER_REGEX ]]
then
  echo "ERROR: Version $(cat VERSION) is not in SemVer format"
  exit 2
fi

# check if version has '-dev'
# if there is, no need to check version
if [[ $(cat VERSION | tr -d '\n' | tail -c 4) == "-dev" ]]
then
  exit 0
fi

# check if the version is already tagged in GitHub repository
for t in $(git tag | cat)
do
  if [[ $t == $(echo v$(cat VERSION | tr -d '\n')) ]]
  then
    echo "ERROR: duplicated tag: $t"
    exit 2
  fi
done