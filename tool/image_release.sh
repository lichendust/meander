#!/bin/sh
mogrify -path images -quality 90 -resize 1400x -format webp images/*.png