#!/bin/bash

rm -f build/*.zip
rm -f build/*.sha512sum

build_dir="build"

set -e

if [ -z $1 ]; then
	echo "please specify a version"
	exit 1;
fi

rm -f $build_dir/*.zip
rm -f $build_dir/*.sha512sum

printf "[packaging]\n"

for f in $build_dir/*; do
	base=meander_$(basename $f)

	echo $base

	name=${base/"_"/"_$1_"}

	mkdir -p $f/license/

	cp -n font/OFL.txt    $f/license/courier_prime.txt
	cp -n license         $f/license/meander.txt
	cp -n readme.fountain $f/readme.fountain

	pushd $f > /dev/null
	zip -r "../$name.zip" * > /dev/null
	popd > /dev/null

	pushd $build_dir > /dev/null
	sha512sum "$name.zip" > "$name.sha512sum"
	popd > /dev/null
done

printf "\n[checksums]\n"

pushd $build_dir > /dev/null
sha512sum -c *.sha512sum
popd > /dev/null