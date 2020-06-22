#! /bin/sh
# Copyright (c) 2004-present Facebook All rights reserved.
# Use of this source code is governed by a BSD-style
# license that can be found in the LICENSE file.


echo ""
echo "------> Downloading magma master to /tmp/magma.zip"
curl -s -L https://github.com/facebookincubator/magma/archive/master.zip --output /tmp/magma.zip


# Copy all entities from inventory
echo ""
echo "------> Extracting inventory ent/ and merging with resourceManager/ent/"
mkdir -p ./inv/
unzip -qq -o /tmp/magma.zip "magma-master/symphony/pkg/ent/*" -d ./inv
ls inv/magma-master/symphony/pkg/ent/
echo "------> Prefixing inventory files with inv_"
for file in $(find ./inv -type f)
do
    if grep -q "Code generated (@generated) by entc, DO NOT EDIT." "$file"
    then
        # echo "Generated file: $file. Removing"
        # rm $file
        :
    else
        mv $file "${file%/*}/inv_"`basename $file`
    fi
done
echo "------> Replacing inventory ent/ imports with local"
find inv/magma-master/symphony/pkg/ent/ -type f -exec sed -i 's/github.com\/facebookincubator\/symphony\/pkg/github.com\/marosmars\/resourceManager/g' {} +
echo "------> Copying local files for inventory ent/"
cp -r inv/magma-master/symphony/pkg/ent/* ./
rm -rf inv
echo "------> Successfully extracted inventory ent/"


# Copy all the dependencies of entities
for folder in authz actions viewer ent-contrib log mysql telemetry
do
    echo ""
    echo "------> Extracting inventory $folder into $folder"
    mkdir -p ../"${folder}"
    unzip -qq -o /tmp/magma.zip "magma-master/symphony/pkg/${folder}/*" -d "../${folder}/"
    ls "../${folder}/magma-master/symphony/pkg/${folder}/"
    echo "------> Prefixing $folder files with inv_"
    for file in $(find ../${folder} -type f)
    do
        mv $file "${file%/*}/inv_"`basename $file`
    done
    echo "------> Replacing $folder imports with local"
    find "../${folder}/magma-master/symphony/pkg/${folder}/" -type f -exec sed -i 's/github.com\/facebookincubator\/symphony\/pkg/github.com\/marosmars\/resourceManager/g' {} +
    cp -r "../${folder}/magma-master/symphony/pkg/${folder}/" ../
    rm -rf "../${folder}/magma-master"
    echo "------> Extracting $folder SUCCESS"
done