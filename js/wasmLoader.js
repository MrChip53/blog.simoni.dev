(function () {
  async function loadGoWasm(wasm, target) {
    if (WebAssembly && !WebAssembly.instantiateStreaming) { // polyfill
      WebAssembly.instantiateStreaming = async (resp, importObject) => {
        const source = await (await resp).arrayBuffer();
        return await WebAssembly.instantiate(source, importObject);
      };
    }

    const go = new Go();

    go.argv = [wasm, target];

    const result = await WebAssembly.instantiateStreaming(fetch(wasm), go.importObject);
    go.run(result.instance);
  }

  function initWasmContainer(container) {
    const wasmUrl = container.dataset.wasmUrl;
    const type = container.dataset.type;

    if (!wasmUrl) {
      console.error('WASM URL not specified for container', container);
      return;
    }
    if (!type) {
      console.error('WASM type not specified for container', container);
      return;
    }

    switch (type) {
      case 'go':
        loadGoWasm(wasmUrl, container.id).catch(err => {
          console.error('Failed to load Go WASM:', err);
        });
        break;
      default:
        console.error('Unsupported WASM type:', type);
    }
  }

  // Initialize on page load
  document.addEventListener('DOMContentLoaded', () => {
    document.querySelectorAll('div[data-wasm-url]').forEach(initWasmContainer);

    document.body.addEventListener('htmx:afterSwap', (event) => {
      event.detail.target.querySelectorAll('div[data-wasm-url]').forEach(initWasmContainer);
    });
  });
})();