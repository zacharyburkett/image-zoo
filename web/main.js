const statusEl = document.getElementById("status");
const seedInput = document.getElementById("seed");
const sizeInput = document.getElementById("size");
const renderButton = document.getElementById("render");

const setStatus = (message) => {
  statusEl.textContent = message;
};

window.setStatus = setStatus;

const go = new Go();
let wasmReady = false;

async function loadWasm() {
  try {
    const result = await WebAssembly.instantiateStreaming(
      fetch("image_zoo.wasm"),
      go.importObject
    );
    wasmReady = true;
    renderButton.disabled = false;
    setStatus("Ready to evolve a pattern.");
    go.run(result.instance);
  } catch (err) {
    setStatus("Wasm load failed. Check console.");
    console.error(err);
  }
}

renderButton.addEventListener("click", () => {
  if (!wasmReady || typeof window.renderImage !== "function") {
    setStatus("Wasm not ready yet.");
    return;
  }
  const seed = Number.parseInt(seedInput.value || "0", 10);
  const size = Number.parseInt(sizeInput.value || "256", 10);
  window.renderImage(seed, size);
});

sizeInput.addEventListener("input", () => {
  const size = Number.parseInt(sizeInput.value || "256", 10);
  document.getElementById("canvas").style.width = `${Math.min(520, size)}px`;
});

loadWasm();
