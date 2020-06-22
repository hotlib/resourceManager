#! /bin/sh

# Cleanup
echo ""
echo "------> Removing all inv_* files"
find . -type f -name 'inv_*' -delete
echo "------> Removing all gen_* files"
find . -type f -name 'gen_*' -delete

go generate ./ent

echo ""
echo "Executing main.go"
go run main.go
