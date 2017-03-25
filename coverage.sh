#!/bin/sh
mkdir -p coverage
go test -coverprofile=coverage/binaryutils.out scummatlas/binaryutils
go test -coverprofile=coverage/condlog.out scummatlas/condlog
go test -coverprofile=coverage/image.out scummatlas/image
go test -coverprofile=coverage/blocks.out scummatlas/blocks
go test -coverprofile=coverage/templates.out scummatlas/templates
go test -coverprofile=coverage/script.out scummatlas/script
go tool cover -html=coverage/main.out -o coverage/main.html
go tool cover -html=coverage/condlog.out -o coverage/condlog.html
go tool cover -html=coverage/templates.out -o coverage/templates.html
go tool cover -html=coverage/script.out -o coverage/script.html
go tool cover -html=coverage/image.out -o coverage/image.html
go tool cover -html=coverage/blocks.out -o coverage/blocks.html
rm coverage/*.out
