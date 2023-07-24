const go = new Go();
WebAssembly.instantiateStreaming(fetch("json.wasm"), go.importObject).then((result) => {
  go.run(result.instance);
});
