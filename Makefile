counters:
	go generate ./...
	tinygo build --no-debug -o index/main.wasm -target wasm examples/countersExample/main.go

pokedex:
	go generate ./...
	tinygo build --no-debug -o index/main.wasm -target wasm examples/pokedex/main.go

clean:
	rm -rf index/main.wasm