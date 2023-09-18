# goFE
A Go frontend framework. Just because.

## Prerequisites
- `go`
- `tinygo`
- `docker compose`

### To run

Firstly, build by doing `make pokedex`. This will write the WebAssembly binary to `index/main.wasm`. Then, to run the nginx server, `docker compose up -d`, and navigate to `http://localhost/`.
