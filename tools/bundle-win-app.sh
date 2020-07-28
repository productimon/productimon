#!/bin/bash

exe_file=$1
out=$2

if [ $# -ne 2 ]
then
  echo "Usage $0 exe_file output_zip"
fi

mkdir productimon-reporter
cp "$exe_file" productimon-reporter/reporter.exe

windeployqt productimon-reporter/reporter.exe > /dev/null || exit 1

# copy all dlls from mingw to the bundled dir
ldd productimon-reporter/reporter.exe | grep "=> /mingw64/" | awk '{print $3}' | xargs -I{} cp {} productimon-reporter

zip -r "$out" productimon-reporter > /dev/null || exit 1
rm -rf productimon-reporter
