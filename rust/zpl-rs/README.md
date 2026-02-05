# zpl-rs

A Rust library for rendering ZPL (Zebra Programming Language) labels to images.

This crate provides bindings to [go-zpl](https://github.com/StirlingMarketingGroup/go-zpl), a native ZPL parser and renderer. Everything runs locally - no external services required.

## Features

- Parse and render ZPL to PNG images
- Support for text, barcodes (Code 128, QR, DataMatrix, PDF417, etc.), and graphics
- Configurable DPI (203, 300, 600)
- Cross-platform: Linux, macOS, Windows (x64 and ARM64)
- Automatic library download at build time

## Installation

Add to your `Cargo.toml`:

```toml
[dependencies]
zpl-rs = "0.1"
```

## Quick Start

```rust
use zpl_rs::render;

fn main() {
    let zpl = r#"^XA
^FO50,50^A0N,30,30^FDHello, World!^FS
^FO50,100^BQN,2,5^FDMA,https://example.com^FS
^XZ"#;

    // Render to PNG bytes
    let png_bytes = render(zpl).expect("Failed to render ZPL");

    // Save to file
    std::fs::write("label.png", &png_bytes).unwrap();
    println!("Saved label.png ({} bytes)", png_bytes.len());
}
```

## Advanced Usage

```rust
use zpl_rs::{render_with_options, RenderOptions, Dpi};

let zpl = "^XA^FO50,50^A0N,30,30^FDHello!^FS^XZ";

let options = RenderOptions::new()
    .dpi(Dpi::Dpi300)
    .size(1200, 1800);  // 4" x 6" at 300 DPI

let png_bytes = render_with_options(zpl, &options).expect("Failed to render");
```

## API

### Functions

- `render(zpl: &str) -> Result<Vec<u8>>` - Render ZPL with default settings (203 DPI)
- `render_bytes(zpl: &[u8]) -> Result<Vec<u8>>` - Render ZPL bytes (non-UTF8 safe)
- `render_with_options(zpl: &str, options: &RenderOptions) -> Result<Vec<u8>>` - Render with custom options
- `render_bytes_with_options(zpl: &[u8], options: &RenderOptions) -> Result<Vec<u8>>` - Render bytes with options

### Types

- `Dpi` - Printer DPI: `Dpi203`, `Dpi300`, `Dpi600`
- `RenderOptions` - Configuration for rendering (DPI, width, height)
- `Error` - Error types: `ParseError`, `RenderError`, `InternalError`

## Supported Platforms

| Platform | Architecture | Status |
|----------|-------------|--------|
| Linux | x86_64 | ✅ |
| Linux | aarch64 | ✅ |
| macOS | x86_64 | ✅ |
| macOS | aarch64 | ✅ |
| Windows | x86_64 | ✅ |
| Windows | aarch64 | ✅ |

## Thread Safety

**Important:** The underlying Go library is not thread-safe for concurrent renders. If you need to render from multiple threads, use a mutex:

```rust
use std::sync::Mutex;
use zpl_rs::render;

static ZPL_MUTEX: Mutex<()> = Mutex::new(());

fn render_safe(zpl: &str) -> zpl_rs::Result<Vec<u8>> {
    let _lock = ZPL_MUTEX.lock().unwrap();
    render(zpl)
}
```

## How It Works

At build time, `zpl-rs` automatically downloads the appropriate prebuilt `libzpl` shared library for your platform from GitHub releases. The library is then linked dynamically.

You can override the version with the `LIBZPL_VERSION` environment variable:
```bash
LIBZPL_VERSION=0.2.0 cargo build
```

## License

MIT
