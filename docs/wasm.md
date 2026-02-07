# WASM App

This directory contains a minimal wasm front-end that renders a single CPPN sample.

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
