#build:
#	GOARCH=wasm GOOS=js go build -o index/main.wasm wasm/main.go

build:
	go generate ./...
	tinygo build --no-debug -o index/main.wasm -target wasm wasm/main.go

clean:
	rm -rf index/main.wasm