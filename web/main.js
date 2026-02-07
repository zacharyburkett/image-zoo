const statusEl = document.getElementById("status");
const seedInput = document.getElementById("seed");
const sizeInput = document.getElementById("size");
const populationInput = document.getElementById("population");
const generationsInput = document.getElementById("generations");
const renderButton = document.getElementById("render");

const setStatus = (message) => {
  statusEl.textContent = message;
};

const setRunning = (running) => {
  renderButton.disabled = running || !wasmReady;
  renderButton.textContent = running ? "Generating…" : "Generate";
};

window.setStatus = setStatus;
window.setRunning = setRunning;

const go = new Go();
let wasmReady = false;

async function loadWasm() {
  try {
    const result = await WebAssembly.instantiateStreaming(
      fetch("image_zoo.wasm"),
      go.importObject
    );
    wasmReady = true;
    setRunning(false);
    setStatus("Ready to evolve a gallery.");
    go.run(result.instance);
  } catch (err) {
    setStatus("Wasm load failed. Check console.");
    console.error(err);
  }
}

renderButton.addEventListener("click", () => {
  if (!wasmReady || typeof window.renderGallery !== "function") {
    setStatus("Wasm not ready yet.");
    return;
  }
  const seed = Number.parseInt(seedInput.value || "0", 10);
  const size = Number.parseInt(sizeInput.value || "192", 10);
  const population = Number.parseInt(populationInput.value || "16", 10);
  const generations = Number.parseInt(generationsInput.value || "8", 10);
  setRunning(true);
  window.renderGallery(seed, size, population, generations);
});

sizeInput.addEventListener("input", () => {
  const size = Number.parseInt(sizeInput.value || "192", 10);
  document.getElementById("canvas").style.width = `${Math.min(640, size)}px`;
});

loadWasm();
