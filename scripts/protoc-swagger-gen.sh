#!/usr/bin/env bash

set -eo pipefail

mkdir -p ./tmp-swagger-gen

cd proto
echo "Generate realio swagger files"
proto_dirs=$(find ./ -path -prune -o -name '*.proto' -print0 | xargs -0 -n1 dirname | sort | uniq)
for dir in $proto_dirs; do
  # generate swagger files (filter query files)
  query_file=$(find "${dir}" -maxdepth 1 \( -name 'query.proto' -o -name 'service.proto' \))
  if [[ ! -z "$query_file" ]]; then
    echo "$query_file"
    buf generate --template buf.gen.swagger.yaml "$query_file"
  fi
done
rm -rf github.com

cd ../third_party/proto
buf mod update
echo "Generate third_party swagger files"

proto_dirs=$(find ./cosmos ./ibc ./multistaking ./evmos -path -prune -o -name '*.proto' -print0 | xargs -0 -n1 dirname | sort | uniq)
for dir in $proto_dirs; do
  # generate swagger files (filter query files)
  query_file=$(find "${dir}" -maxdepth 1 \( -name 'query.proto' -o -name 'service.proto' \))
  if [[ ! -z "$query_file" ]]; then
    echo "$query_file"
    buf generate --template buf.gen.swagger.yaml "$query_file"
  fi
done

cd ../..

mv third_party/tmp-swagger-gen/* tmp-swagger-gen
echo "Combine swagger files"
# combine swagger files
# uses nodejs package `swagger-combine`.
# all the individual swagger files need to be configured in `config.json` for merging
swagger-combine ./client/docs/config.json -o ./client/docs/swagger-ui/swagger.yaml -f yaml --continueOnConflictingPaths true --includeDefinitions true

# clean swagger files
rm -rf ./tmp-swagger-gen
rm -rf ./third_party

echo "Update statik data"
install_statik() {
  go install github.com/rakyll/statik@v0.1.7
}
install_statik

# generate binary for static server
statik -f -src=./client/docs/swagger-ui -dest=./client/docs
