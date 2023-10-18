#!/usr/bin/env bash

output="$(python3 -m pip list | grep -F graphviz)"
if ! [[ -n $output ]]; then
  python3 -m pip install graphviz
fi

python3 toPng.py

ffmpeg -framerate 3 -i out/%010d.png -s 350x250 -pix_fmt yuv420p output.mp4
