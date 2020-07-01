#! /bin/sh
# Copyright (c) 2004-present Facebook All rights reserved.
# Use of this source code is governed by a BSD-style
# license that can be found in the LICENSE file.

echo ""
echo "------> Downloading magma master to /tmp/magma.zip"
MAGMA_VERSION="7bd82a1ba7000415cd5d394ff026ab66ed4c5987"
if [ -f /tmp/magma.zip ]; then
    echo "------> Skipping magma download, archive already present in /tmp/magma.zip, Remove it if you need redownload !"
    sleep 1
else
    curl -s -L https://github.com/facebookincubator/magma/archive/"${MAGMA_VERSION}".zip --output /tmp/magma.zip
fi

# Copy all entities from inventory
echo ""
echo "------> Extracting inventory ent/ and merging with resourceManager/ent/"
mkdir -p ./inv/
unzip -qq -o /tmp/magma.zip "magma-${MAGMA_VERSION}/symphony/pkg/ent/*" -d ./inv
ls inv/magma-"${MAGMA_VERSION}"/symphony/pkg/ent/
echo "------> Prefixing inventory files with inv_"
for file in $(find ./inv -type f); do
    if grep -q "Code generated (@generated) by entc, DO NOT EDIT." "$file"; then
        # echo "Generated file: $file. Removing"
        # rm $file
        :
    else
        mv $file "${file%/*}/inv_"$(basename $file)
    fi
done
echo "------> Replacing inventory ent/ imports with local"
find inv/magma-"${MAGMA_VERSION}"/symphony/pkg/ent/ -type f -exec sed -i 's/github.com\/facebookincubator\/symphony\/pkg/github.com\/marosmars\/resourceManager/g' {} +
echo "------> Copying local files for inventory ent/"
cp -r inv/magma-"${MAGMA_VERSION}"/symphony/pkg/ent/* ./
rm -rf inv
echo "------> Successfully extracted inventory ent/"

# Copy all the dependencies of entities
for folder in authz actions viewer ent-contrib log mysql telemetry ctxgroup ctxutil server; do
    echo ""
    echo "------> Extracting inventory $folder into $folder"
    mkdir -p ../"${folder}"
    unzip -qq -o /tmp/magma.zip "magma-${MAGMA_VERSION}/symphony/pkg/${folder}/*" -d "../${folder}/"
    ls "../${folder}/magma-${MAGMA_VERSION}/symphony/pkg/${folder}/"
    echo "------> Prefixing $folder files with inv_"
    for file in $(find ../${folder} -type f); do
        mv $file "${file%/*}/inv_"$(basename $file)
    done
    echo "------> Replacing $folder imports with local"
    find "../${folder}/magma-${MAGMA_VERSION}/symphony/pkg/${folder}/" -type f -exec sed -i 's/github.com\/facebookincubator\/symphony\/pkg/github.com\/marosmars\/resourceManager/g' {} +
    cp -r "../${folder}/magma-${MAGMA_VERSION}/symphony/pkg/${folder}/" ../
    rm -rf "../${folder}/magma-${MAGMA_VERSION}"
    echo "------> Removing all test files"
    find ../"${folder}" -type f -name '*_test.go' -delete
    echo "------> Extracting $folder SUCCESS"
done

# Copy all the dependencies of graphql

# TODO uncomment for custom directive support, directives are custom extensions to the graphql schema such as number range restrictions, see inventory/**/symphony.graphql for custom directives
# for folder in plugin directive; do
for folder in plugin; do
    echo ""
    echo "------> Extracting inventory $folder into $folder"
    mkdir -p ../graph/graphql/"${folder}"
    unzip -qq -o /tmp/magma.zip "magma-${MAGMA_VERSION}/symphony/graph/graphql/${folder}/*" -d "../graph/graphql/${folder}/"
    ls "../graph/graphql/${folder}/magma-${MAGMA_VERSION}/symphony/graph/graphql/${folder}/"
    echo "------> Prefixing ${folder} files with inv_"
    for file in $(find ../graph/graphql/${folder} -type f); do
        mv $file "${file%/*}/inv_"$(basename $file)
    done
    echo "------> Replacing ${folder} imports with local"
    find "../graph/graphql/${folder}/magma-${MAGMA_VERSION}/symphony/graph/graphql/${folder}/" -type f -exec sed -i 's/github.com\/facebookincubator\/symphony\/pkg/github.com\/marosmars\/resourceManager/g' {} +
    find "../graph/graphql/${folder}/magma-${MAGMA_VERSION}/symphony/graph/graphql/${folder}/" -type f -exec sed -i 's/github.com\/facebookincubator\/symphony/github.com\/marosmars\/resourceManager/g' {} +
    echo "------> Adding ent import in plugin"
    sed -i '1s/^/{{ reserveImport \"github.com\/marosmars\/resourceManager\/ent" }}\n/' "../graph/graphql/plugin/magma-${MAGMA_VERSION}/symphony/graph/graphql/plugin/txgen/inv_txgen.gotpl"
    cp -r "../graph/graphql/${folder}/magma-${MAGMA_VERSION}/symphony/graph/graphql/${folder}/" ../graph/graphql/
    rm -rf "../graph/graphql/${folder}/magma-${MAGMA_VERSION}"
    echo "------> Removing all test files"
    find ../graph/graphql/"${folder}" -type f -name '*_test.go' -delete
    echo "------> Extracting ${folder} SUCCESS"
done
