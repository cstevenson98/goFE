counters:
	go generate ./...
	tinygo build --no-debug -o index/main.wasm -target wasm examples/countersExample/main.go

pokedex:
	go generate ./...
	tinygo build --no-debug -o index/main.wasm -target wasm examples/pokedex/main.go

router:
	go generate ./...
	tinygo build --no-debug -o index/main.wasm -target wasm examples/routerExample/main.go

test-gox:
	go test ./pkg/goFE/gox

clean:
	rm -rf index/main.wasm