# zpl-rs

A Rust library for rendering ZPL (Zebra Programming Language) labels to images.

This crate provides bindings to [go-zpl](https://github.com/StirlingMarketingGroup/go-zpl), a native ZPL parser and renderer. Everything runs locally - no external services required.

## Features

- Parse and render ZPL to PNG images
- Support for text, barcodes (Code 128, QR, DataMatrix, PDF417, etc.), and graphics
- Configurable DPI (203, 300, 600)
- Cross-platform: Linux, macOS, Windows (x64 and ARM64)
- Automatic library download at build time, or an offline build via `LIBZPL_PATH`

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

## How it works

At build time the crate stages the prebuilt `libzpl` shared library for the target into its `OUT_DIR` (downloaded from GitHub releases, or copied from `LIBZPL_PATH`) and links it dynamically (on Windows through `raw-dylib`, so no import library is needed). `cargo run` / `cargo test` in a downstream project need no extra setup: Cargo adds `OUT_DIR` library paths to the dynamic-library search path for those commands. Distributing the built binary is where it breaks: the library is not next to the binary, and the rpaths this crate sets do not apply to a downstream binary (Cargo applies a dependency's `rustc-link-arg` only to that dependency's own targets).

## Environment variables

| Var | Effect |
|-----|--------|
| `LIBZPL_VERSION` | libzpl release to download (default: latest GitHub release). Changing it replaces the staged library on the next build. |
| `LIBZPL_PATH` | path to an already-built libzpl shared library for the target; staged instead of downloading (offline builds, a locally built `cmd/libzpl`, CI caches). Wins over `LIBZPL_VERSION`; no network access. |
| `LIBZPL_COPY_TO` | directory to also copy the staged library into (created if missing). Opt-in staging hook for bundlers; writes outside `OUT_DIR` on purpose. |

## Build metadata for dependents

The crate declares `links = "zpl"`, so the build script of a crate that depends on `zpl-rs` **directly** (Cargo passes metadata to immediate dependents only) gets:

| Variable | Value |
|----------|-------|
| `DEP_ZPL_LIB_DIR` | absolute directory that holds the staged library |
| `DEP_ZPL_LIB_PATH` | absolute path of the staged library file |
| `DEP_ZPL_LIB_NAME` | file name (`libzpl.dylib` / `libzpl.so` / `zpl.dll`) |
| `DEP_ZPL_VERSION` | the release version, or the literal `file` when `LIBZPL_PATH` was used |

This is the supported way to locate the library; never scan `target/`.

## Bundling for distribution

Two things have to happen:

1. The library file must end up where the OS loader looks at runtime — next to the executable, or a directory an rpath names.
2. On macOS/Linux the binary needs rpaths that point there, and those must be emitted by *your* build script.

### rpaths and staging from your own `build.rs`

`cargo run` / `cargo test` still work without this; it only matters for a relocated binary. This is a complete `build.rs` for the application crate. The first half sets the rpaths; the second half stages the library into `lib/` next to your `Cargo.toml` for the bundler (drop it if you stage with `LIBZPL_COPY_TO` instead):

```rust
use std::env;
use std::fs;
use std::path::PathBuf;

fn main() {
    let os = env::var("CARGO_CFG_TARGET_OS").unwrap();
    match os.as_str() {
        "macos" => {
            println!("cargo:rustc-link-arg=-Wl,-rpath,@executable_path");
            println!("cargo:rustc-link-arg=-Wl,-rpath,@executable_path/../Frameworks");
        }
        "linux" => {
            println!("cargo:rustc-link-arg=-Wl,-rpath,$ORIGIN");
            println!("cargo:rustc-link-arg=-Wl,-rpath,$ORIGIN/../lib");
        }
        // Windows: the DLL must sit next to the .exe (or on PATH).
        _ => {}
    }

    // Stage the library for the bundler. DEP_ZPL_* exist only in a crate that depends on
    // zpl-rs directly. Cargo discourages build scripts writing outside OUT_DIR; this is the
    // pragmatic route app bundlers need.
    let lib_path = env::var("DEP_ZPL_LIB_PATH")
        .expect("DEP_ZPL_LIB_PATH unset; zpl-rs must be a direct dependency");
    let lib_name = env::var("DEP_ZPL_LIB_NAME").unwrap();
    let dest_dir = PathBuf::from(env::var("CARGO_MANIFEST_DIR").unwrap()).join("lib");
    fs::create_dir_all(&dest_dir).unwrap();
    let dest = dest_dir.join(&lib_name);
    // Copy only when missing or different. Cargo takes the build script's reference time
    // at invocation, so a file this script writes and also watches reads as changed on
    // the next build; an unconditional copy would rerun this script forever.
    let staged = fs::read(&lib_path).unwrap();
    if fs::read(&dest).ok().as_deref() != Some(staged.as_slice()) {
        fs::copy(&lib_path, &dest).unwrap();
    }
    // Re-stage if the copy is deleted or replaced (a clean step, a fresh checkout);
    // without this Cargo considers the script fresh and never recreates it.
    println!("cargo:rerun-if-changed={}", dest.display());
}
```

### Staging with `LIBZPL_COPY_TO` instead

If you would rather not copy from `build.rs`, set `LIBZPL_COPY_TO` and `zpl-rs` does the copy itself:

```bash
LIBZPL_COPY_TO=$PWD/lib cargo build --release
```

or in `.cargo/config.toml`:

```toml
[env]
LIBZPL_COPY_TO = { value = "lib", relative = true }
```

Relative values resolve differently in the two forms: a raw environment value is taken relative to the `zpl-rs` package directory (the build script's working directory), while `[env]` with `relative = true` resolves against the directory that holds `.cargo/` and reaches the build script as an absolute path. Prefer an absolute path or the `[env]` form.

The staging directory is the bundler's **input**, not "next to the binary": Cargo puts the binary under `<target-dir>/[<target-triple>/]<profile>/`, so for a bare binary locate the actual executable directory and copy the library beside it yourself (or point `LIBZPL_COPY_TO` at that directory with an absolute path).

### Bundler notes

- **Tauri**: `bundle.macOS.frameworks: ["lib/libzpl.dylib"]` (relative to `src-tauri/`) puts the dylib in `Contents/Frameworks/` — the `@executable_path/../Frameworks` rpath above is what makes that work. On Linux, `LD_LIBRARY_PATH` must include the directory holding `libzpl.so` when running `tauri build` so linuxdeploy bundles it into the AppImage. On Windows use the object form of `bundle.resources` so the DLL lands in the resource root next to the exe: `"resources": { "lib/zpl.dll": "zpl.dll" }` (the array form keeps the `lib/` subdirectory, which the Windows loader does not search). The staging step above is what produces `lib/` for these configs.
- **cargo-bundle / other**: same idea — stage from `DEP_ZPL_LIB_PATH` or `LIBZPL_COPY_TO`, tell the bundler about the file, set rpaths.
- **Bare binary**: ship the library next to the executable and use the `$ORIGIN` / `@executable_path` rpaths.

## License

MIT
