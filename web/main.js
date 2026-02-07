const statusEl = document.getElementById("status");
const seedInput = document.getElementById("seed");
const populationInput = document.getElementById("population");
const generationsInput = document.getElementById("generations");
const modeSelect = document.getElementById("mode");
const renderButton = document.getElementById("render");
const stopButton = document.getElementById("stop");
const presetSelect = document.getElementById("preset");
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

const weightInputs = {
  entropy: document.getElementById("w-entropy"),
  edge: document.getElementById("w-edge"),
  fine: document.getElementById("w-fine"),
  variance: document.getElementById("w-variance"),
  symmetry: document.getElementById("w-symmetry"),
  color: document.getElementById("w-color"),
  highfreq: document.getElementById("w-highfreq"),
  novelty: document.getElementById("w-novelty"),
};

const weightValues = {
  entropy: document.getElementById("w-entropy-value"),
  edge: document.getElementById("w-edge-value"),
  fine: document.getElementById("w-fine-value"),
  variance: document.getElementById("w-variance-value"),
  symmetry: document.getElementById("w-symmetry-value"),
  color: document.getElementById("w-color-value"),
  highfreq: document.getElementById("w-highfreq-value"),
  novelty: document.getElementById("w-novelty-value"),
};

const presets = {
  balanced: {
    entropy: 0.2,
    edge: 0.35,
    fine: 0.15,
    variance: 0.2,
    symmetry: 0.1,
    color: 0.15,
    highfreq: 0.25,
    novelty: 0.2,
  },
  organic: {
    entropy: 0.15,
    edge: 0.2,
    fine: 0.2,
    variance: 0.35,
    symmetry: 0.05,
    color: 0.15,
    highfreq: 0.15,
    novelty: 0.25,
  },
  geometric: {
    entropy: 0.1,
    edge: 0.45,
    fine: 0.1,
    variance: 0.15,
    symmetry: 0.25,
    color: 0.1,
    highfreq: 0.1,
    novelty: 0.15,
  },
  symmetric: {
    entropy: 0.1,
    edge: 0.2,
    fine: 0.05,
    variance: 0.1,
    symmetry: 0.45,
    color: 0.1,
    highfreq: 0.1,
    novelty: 0.1,
  },
  psychedelic: {
    entropy: 0.25,
    edge: 0.55,
    fine: 0.55,
    variance: 0.2,
    symmetry: 0.05,
    color: 0.2,
    highfreq: 0.05,
    novelty: 0.3,
  },
};

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

const updateWeightDisplay = (key) => {
  weightValues[key].textContent = Number.parseFloat(
    weightInputs[key].value
  ).toFixed(2);
};

const setWeights = (weights) => {
  Object.keys(weightInputs).forEach((key) => {
    const value = weights[key];
    if (typeof value === "number") {
      weightInputs[key].value = value.toFixed(2);
      updateWeightDisplay(key);
    }
  });
};

const readWeights = () => ({
  entropy: Number.parseFloat(weightInputs.entropy.value),
  edge: Number.parseFloat(weightInputs.edge.value),
  fine: Number.parseFloat(weightInputs.fine.value),
  variance: Number.parseFloat(weightInputs.variance.value),
  symmetry: Number.parseFloat(weightInputs.symmetry.value),
  color: Number.parseFloat(weightInputs.color.value),
  highfreq: Number.parseFloat(weightInputs.highfreq.value),
  novelty: Number.parseFloat(weightInputs.novelty.value),
});

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

window.updateDetail = (
  width,
  height,
  pixels,
  fitness,
  nodes,
  conns,
  hidden,
  outputs,
  summary
) => {
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

Object.keys(weightInputs).forEach((key) => {
  weightInputs[key].addEventListener("input", () => {
    updateWeightDisplay(key);
    presetSelect.value = "custom";
  });
  updateWeightDisplay(key);
});

presetSelect.addEventListener("change", () => {
  const preset = presets[presetSelect.value];
  if (preset) {
    setWeights(preset);
  }
});

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
  const population = Number.parseInt(populationInput.value || "16", 10);
  const generations = Number.parseInt(generationsInput.value || "8", 10);
  const colorMode = modeSelect.value === "color" ? 1 : 0;
  const weights = readWeights();
  setRunning(true);
  window.renderGallery(
    seed,
    population,
    generations,
    colorMode,
    weights.entropy,
    weights.edge,
    weights.fine,
    weights.variance,
    weights.symmetry,
    weights.color,
    weights.highfreq,
    weights.novelty
  );
});

stopButton.addEventListener("click", () => {
  if (typeof window.stopEvolution === "function") {
    window.stopEvolution();
  }
});

modalBackdrop.addEventListener("click", closeModal);
modalClose.addEventListener("click", closeModal);

setWeights(presets.balanced);

loadWasm();
