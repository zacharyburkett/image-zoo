const statusEl = document.getElementById("status");
const seedInput = document.getElementById("seed");
const sizeInput = document.getElementById("size");
const populationInput = document.getElementById("population");
const generationsInput = document.getElementById("generations");
const modeSelect = document.getElementById("mode");
const renderButton = document.getElementById("render");
const stopButton = document.getElementById("stop");
const galleryEl = document.getElementById("gallery");
const modal = document.getElementById("modal");
const modalBackdrop = document.getElementById("modal-backdrop");
const modalClose = document.getElementById("modal-close");
const detailCanvas = document.getElementById("detail-canvas");
const detailFitness = document.getElementById("detail-fitness");
const detailNodes = document.getElementById("detail-nodes");
const detailConnections = document.getElementById("detail-connections");
const detailHidden = document.getElementById("detail-hidden");
const detailOutputs = document.getElementById("detail-outputs");
const detailSummary = document.getElementById("detail-summary");

let tiles = [];

const setStatus = (message) => {
  statusEl.textContent = message;
};

const setRunning = (running) => {
  renderButton.disabled = running || !wasmReady;
  stopButton.disabled = !running;
  renderButton.textContent = running ? "Generating…" : "Generate";
};

const openModal = () => {
  modal.classList.add("open");
  modal.setAttribute("aria-hidden", "false");
};

const closeModal = () => {
  modal.classList.remove("open");
  modal.setAttribute("aria-hidden", "true");
};

window.setStatus = setStatus;
window.setRunning = setRunning;

window.prepareGallery = (count, tileSize) => {
  tiles = [];
  galleryEl.innerHTML = "";
  galleryEl.style.setProperty("--tile-size", `${tileSize}px`);
  for (let i = 0; i < count; i += 1) {
    const tile = document.createElement("div");
    tile.className = "tile";
    tile.dataset.index = String(i);
    const canvas = document.createElement("canvas");
    canvas.width = tileSize;
    canvas.height = tileSize;
    tile.appendChild(canvas);
    tile.addEventListener("click", () => {
      if (typeof window.renderDetail === "function") {
        const detailSize = Math.max(512, tileSize * 2);
        window.renderDetail(i, detailSize);
        openModal();
      }
    });
    galleryEl.appendChild(tile);
    tiles.push(canvas);
  }
};

window.updateTile = (index, width, height, pixels) => {
  const canvas = tiles[index];
  if (!canvas) {
    return;
  }
  canvas.width = width;
  canvas.height = height;
  const ctx = canvas.getContext("2d");
  const imageData = new ImageData(pixels, width, height);
  ctx.putImageData(imageData, 0, 0);
};

window.updateDetail = (width, height, pixels, fitness, nodes, conns, hidden, outputs, summary) => {
  detailCanvas.width = width;
  detailCanvas.height = height;
  const ctx = detailCanvas.getContext("2d");
  const imageData = new ImageData(pixels, width, height);
  ctx.putImageData(imageData, 0, 0);
  detailFitness.textContent = fitness.toFixed(4);
  detailNodes.textContent = nodes;
  detailConnections.textContent = conns;
  detailHidden.textContent = hidden;
  detailOutputs.textContent = outputs;
  detailSummary.textContent = summary;
};

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
  const colorMode = modeSelect.value === "color" ? 1 : 0;
  setRunning(true);
  window.renderGallery(seed, size, population, generations, colorMode);
});

stopButton.addEventListener("click", () => {
  if (typeof window.stopEvolution === "function") {
    window.stopEvolution();
  }
});

modalBackdrop.addEventListener("click", closeModal);
modalClose.addEventListener("click", closeModal);

loadWasm();
