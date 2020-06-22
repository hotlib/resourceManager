#! /bin/sh
# Copyright (c) 2004-present Facebook All rights reserved.
# Use of this source code is governed by a BSD-style
# license that can be found in the LICENSE file.


echo ""
echo "------> Renaming generated files"
for file in $(find . -type f)
do
    if grep -q "Code generated (@generated) by entc, DO NOT EDIT." "$file"
    then
        # Do not rename yourself
        if [ $file = "./generate.go" ]
        then
            continue
        fi
        if [ $file = "./pull_inventory.sh" ]
        then
            continue
        fi
        if [ $file = "./rename_generated_files.sh" ]
        then
            continue
        fi
        mv $file "${file%/*}/gen_"`basename $file`
    fi
done
