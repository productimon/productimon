#!/bin/bash

rc_file=$1
src_img=$2
out=$3

if [ $# -ne 3 ]
then
  echo "Usage $0 rc_file source_image output_file"
  exit 1
fi


magick convert "$src_img" -define icon:auto-resize="256,128,96,64,48,32,16" logo.ico || exit $?

windres "$rc_file" -O coff -o "$out" || exit $?
