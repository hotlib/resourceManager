#! /bin/sh

# Cleanup
echo ""
echo "------> Removing all inv_* files"
find . -type f -name 'inv_*' -delete
echo "------> Removing all gen_* files"
find . -type f -name 'gen_*' -delete

go generate ./ent
go generate ./graph/graphql

echo ""
echo "------> Building"

go build

echo ""
echo "------> Testing"

go test ./pools/...
go test ./graph/...
