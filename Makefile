buildCounters:
	go generate ./...
	tinygo build --no-debug -o index/main.wasm -target wasm examples/countersExample/main.go

clean:
	rm -rf index/main.wasm