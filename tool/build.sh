#!/bin/bash

set -e

build_dir="build"
version=$(grep 'const VERSION' source/*.go | awk -F"[ \"]+" '/VERSION/{print $4}')

# START
rm -rf $build_dir/*

printf "[building]\n"

build() {
	echo "$1 $2"
	name="$1_$2"
	ext=""

	if [ $1 = 'darwin' ]; then
		name="macos_"
		if [ $2 = 'arm64' ]; then
			name+="silicon"
		else
			name+="intel"
		fi
	fi
	if [ $1 = 'windows' ]; then
		ext=".exe"
	fi

	mkdir -p "$build_dir/$name"
	env GOOS="$1" GOARCH="$2" go build -ldflags "-s -w" -trimpath -o "$build_dir/$name/meander$ext" ./source
}

build windows amd64
build linux   amd64
build linux   arm64
build darwin  amd64
build darwin  arm64
build freebsd amd64

printf "\n[packaging]\n"

for f in $build_dir/*; do
	base=$(basename $f)

	echo $base

	name=meander_${base}_$version

	cp -n license      $f/license.txt
	cp -n readme.md    $f/readme.txt

	pushd $f > /dev/null
	zip -r "../$name.zip" * > /dev/null
	popd > /dev/null

	pushd $build_dir > /dev/null
	sha512sum "$name.zip" > "$name.sha512sum"
	popd > /dev/null
done

printf "\n[checksums]\n"

pushd $build_dir > /dev/null
sha512sum -c *.sha512sum | column -t
popd > /dev/null