test:
	go test scummatlas
	go test scummatlas/binaryutils
	go test scummatlas/condlog
	go test scummatlas/image
	go test scummatlas/script
	go test scummatlas/templates
	go test scummatlas/blocks

coverage:
	./coverage.sh
