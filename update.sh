#!/bin/bash
THE_TAG=`git tag -l | tail -1`
THE_VERSION=`echo $THE_TAG | sed -E 's#v##'`
THE_FILENAME="hellocontest_${THE_VERSION}_amd64.deb"
echo "Updating to the latest version $THE_VERSION ($THE_TAG): $THE_FILENAME"
wget https://github.com/ftl/hellocontest/releases/download/$THE_TAG/$THE_FILENAME
sudo apt install ./$THE_FILENAME
