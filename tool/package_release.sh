#!/bin/bash

set -e

if [ -z $1 ]; then
	echo "please specify a version"
	exit 1;
fi

rm build/*.zip

for f in build/*; do
	name=$(basename $f)

	name=${name/"_"/"_$1_"}

	mkdir -p $f/license/

	cp -n font/OFL.txt $f/license/courier_prime_license.txt
	cp -n license      $f/license/meander_license.txt
	cp -n readme.md    $f/readme.txt

	pushd $f > /dev/null
	zip -r "../$name.zip" * > /dev/null
	popd > /dev/null
done