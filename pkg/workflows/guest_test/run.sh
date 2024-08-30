#! /bin/bash
GOOS=wasip1 GOARCH=wasm go build -o main.wasm main.go
go build -o main main.go &
wasm-opt --enable-bulk-memory -Os main.wasm -o main.os.wasm &
wasm-opt --enable-bulk-memory -Oz main.wasm -o main.oz.wasm &
wasm-opt --enable-bulk-memory -O3 main.wasm -o main.o3.wasm &

wait

du -h *.wasm main | cut -d "M" -f 1

rm main *.wasm