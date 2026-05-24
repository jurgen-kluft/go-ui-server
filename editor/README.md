# macOS 2D UI Editor Build Pipeline (Apple Silicon)

This project features a native, hardware-accelerated 2D UI Editor framework built in Go using **Dear ImGui** via the Metal graphics API backend. This repository contains the standalone automation script used to compile, optimize, and bundle the Go application into a standard macOS `.app` package.

## 📋 Prerequisites

Before running the automated build script, ensure your Apple Silicon Mac has the following tools installed and configured:

1. **Go Toolchain**: Minimum version 1.18+ configured for ARM64 architectures.
2. **Xcode Command Line Tools**: Required for the standard C compiler (`clang`) to build the internal C++ modules of Dear ImGui (`CGO_ENABLED=1`). Install it via terminal:
   ```bash
   xcode-select --install
   ```
3. **Application Asset**: An image file (e.g., `icon.png`, minimum `1024x1024` resolution recommended) to auto-generate the official macOS application bundle icon.

---

## 🚀 Script Setup

1. Save the automation script as `build_mac_app.sh` in the root folder of your Go project.
2. Grant execution permissions to the script via your terminal:
   ```bash
   chmod +x build_mac_app.sh
   ```

---

## 🛠️ Usage Instructions

The build engine supports two pipelines depending on whether you want to automatically process a source image or use a pre-compiled icon file.

### Option A: Auto-Compile App Icon From PNG (Recommended)
Pass the local path of your target image file as the first parameter. The script will tap into macOS native `sips` and `iconutil` layers to create the multi-resolution `.icns` file automatically:

```bash
./build_mac_app.sh path/to/your_icon_source.png
```

### Option B: Use an Existing `.icns` File
If you already possess a completed `icon.icns` asset file inside your project root, execute the script directly without parameters:

```bash
./build_mac_app.sh
```

---

## 📦 Output Architecture

Upon a successful build layout, the engine produces a standalone **`UIEditor.app`** bundle folder in your project root matching the native macOS deployment specifications:

```text
UIEditor.app/
└── Contents/
    ├── Info.plist          # Configures high-DPI Retina layers & OS execution paths
    ├── MacOS/
    │   └── uieditor        # Stripped, optimized native arm64 Go binary
    └── Resources/
        └── AppIcon.icns    # Embedded application icons
```

To boot your newly generated application instance directly from the workspace terminal environment, execute:

```bash
open UIEditor.app
```

---

## ⚙️ Compilation Highlights
* **Native Architecture Focus**: Blocks standard Intel translation emulation overhead by locking compiler flags exclusively onto Apple Silicon (`GOARCH=arm64`, `-arch arm64`).
* **Production Optimization**: Passes heavy vectorization and loop optimizations (`-O3`) to the native Objective-C/C++ wrapping code blocks while completely stripping debug structures (`-ldflags="-s -w"`) to significantly reduce disk and memory size.
* **Retina Asset Scaling**: Automatically attaches high-resolution graphics hints (`NSHighResolutionCapable`) inside the deployed `Info.plist` layout so UI rendering lines, canvas grids, and text arrays present perfectly sharp edges on modern screens.
