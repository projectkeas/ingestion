#!/bin/bash

go list -u -f '{{if (and (not (or .Main .Indirect)) .Update)}}{{.Path}}@{{.Update.Version}}{{end}}' -m all | xargs go get -u 

go mod tidy