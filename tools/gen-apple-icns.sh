#!/bin/bash

src_img=$1
out=$2

if [ $# -ne 2 ]
then
  echo "Usage $0 source_image output_file"
  exit 1
fi

sizes=(16 32 128 256 512)

mkdir icons.iconset

for size in "${sizes[@]}"
do
  double_size=$(($size*2))
  sips -z "$size" "$size" "$src_img" --out "icons.iconset/icon_$size"x"$size.png" > /dev/null || exit $?
  sips -z "$double_size" "$double_size" "$src_img" --out "icons.iconset/icon_$size"x"$size@2x.png" > /dev/null || exit $?
done
iconutil -c icns -o "$out" icons.iconset
rm -R icons.iconset
