counters:
	go generate ./...
	tinygo build --no-debug -o index/main.wasm -target wasm examples/countersExample/main.go

pokedex:
	go generate ./...
	tinygo build --no-debug -o index/main.wasm -target wasm examples/pokedex/main.go

router:
	go generate ./...
	tinygo build --no-debug -o index/main.wasm -target wasm examples/routerExample/main.go

fetch:
	go generate ./...
	tinygo build --no-debug -o index/main.wasm -target wasm examples/fetchExample/main.go

api:
	go generate ./...
	tinygo build --no-debug -o index/main.wasm -target wasm examples/apiExample/main.go

anthropic-agent:
	go generate ./...
	tinygo build --no-debug -o index/main.wasm -target wasm examples/anthropicAgentExample/main.go

webgpu:
	go generate ./...
	tinygo build --no-debug -o index/main.wasm -target wasm examples/webgpuExample/main.go

clean:
	rm -rf index/main.wasm