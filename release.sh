#!/usr/bin/env sh

version="v$(grep version shioriUpdater.go | grep -oE '[0-9]\.[0-9]\.[0-9]')"
if [ $? -eq 1 ]; then
  echo "error: failed to detect the version" 1>&2
  exit 1
fi

latest_version=$(git describe --tags --abbrev=0)

if [ ${version} == ${latest_version} ]; then
  echo "error: this version is the same as the latest one" 1>&2
  exit 1
fi

GOOS=windows GOARCH=386 go build -o pkg/shioriupdater_windows_386.exe

ghr "${version}" pkg/
