# WASM App

This directory contains a minimal wasm front-end that evolves a CPPN population and renders a gallery.

## Build
```
./scripts/build_wasm.sh
```

## Serve
Use any static server from the repository root, then open `web/index.html`.
Example:
```
python3 -m http.server 8080
```

Then visit `http://localhost:8080/web/` in a browser.

## Notes
- The wasm bundle is `web/image_zoo.wasm`.
- The wasm runtime shim is `web/wasm_exec.js`.
- The UI lets you control seed, population size, and generations.
- The gallery updates each generation with a progress indicator.
- You can stop a run early, toggle grayscale/color output, and click tiles for details.
