// Copyright (c) 2004-present Facebook All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ent

//go:generate ./pull_inventory.sh

//go:generate echo ""
//go:generate echo "------> Generating ent.go entities from ent/schema folder"
//go:generate go run github.com/facebookincubator/ent/cmd/entc generate --storage=sql --template ./template --template ../ent-contrib/entgqlgen/template --header "// Code generated (@generated) by entc, DO NOT EDIT." ./schema
//go:generate echo "------> Generating finished"

//go:generate ./rename_generated_files.sh
