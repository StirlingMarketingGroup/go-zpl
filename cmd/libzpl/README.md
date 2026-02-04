# libzpl - C Shared Library for ZPL Rendering

A C-compatible shared library that exposes go-zpl's rendering capabilities for use from other languages (Rust, C, C++, Python, etc.).

## Building

```bash
# Build for current platform
make

# Build for all platforms
make all

# Build for specific platform
make linux      # libzpl-linux-amd64.so
make darwin     # libzpl.dylib (universal binary)
make windows    # libzpl.dll (requires mingw-w64)
```

Output goes to `../../dist/`.

## API

```c
// Render ZPL to PNG
// Returns: 0 on success, negative on error
int zpl_render_png(
    char* zpl_data,    // ZPL content (not null-terminated)
    int zpl_len,       // Length of ZPL data
    int dpi,           // 203, 300, or 600
    int width,         // Label width in dots (0 = auto)
    int height,        // Label height in dots (0 = auto)
    char** png_out,    // Output: PNG data (caller must free)
    int* png_len       // Output: PNG length
);

// Simpler version with defaults (203 DPI, auto dimensions)
int zpl_render_png_simple(
    char* zpl_data,
    int zpl_len,
    char** png_out,
    int* png_len
);

// Free memory allocated by render functions
void zpl_free(char* ptr);
```

### Error Codes

| Code | Meaning |
|------|---------|
| 0 | Success |
| -1 | ZPL parse error |
| -2 | Render error |
| -3 | Internal error |

## Rust Integration

See `rust_example.rs` for complete Rust FFI bindings. Basic usage:

```rust
#[link(name = "zpl")]
extern "C" {
    fn zpl_render_png_simple(
        zpl_data: *const c_char,
        zpl_len: c_int,
        png_out: *mut *mut c_char,
        png_len: *mut c_int,
    ) -> c_int;
    fn zpl_free(ptr: *mut c_char);
}

pub fn render_zpl(zpl: &[u8]) -> Result<Vec<u8>, i32> {
    let mut png_ptr: *mut c_char = std::ptr::null_mut();
    let mut png_len: c_int = 0;

    let result = unsafe {
        zpl_render_png_simple(
            zpl.as_ptr() as *const c_char,
            zpl.len() as c_int,
            &mut png_ptr,
            &mut png_len,
        )
    };

    if result != 0 {
        return Err(result);
    }

    let png_data = unsafe {
        let slice = std::slice::from_raw_parts(png_ptr as *const u8, png_len as usize);
        let owned = slice.to_vec();
        zpl_free(png_ptr);
        owned
    };

    Ok(png_data)
}
```

### Cargo build.rs Example

```rust
fn main() {
    // Tell cargo where to find the library
    println!("cargo:rustc-link-search=native=/path/to/libzpl");
    println!("cargo:rustc-link-lib=dylib=zpl");

    // On macOS, set rpath
    #[cfg(target_os = "macos")]
    println!("cargo:rustc-link-arg=-Wl,-rpath,@executable_path");
}
```

## C Example

```c
#include "libzpl.h"
#include <stdio.h>
#include <string.h>

int main() {
    const char* zpl = "^XA^FO50,50^A0N,30,30^FDHello!^FS^XZ";
    char* png = NULL;
    int len = 0;

    if (zpl_render_png_simple((char*)zpl, strlen(zpl), &png, &len) == 0) {
        FILE* f = fopen("output.png", "wb");
        fwrite(png, 1, len, f);
        fclose(f);
        zpl_free(png);
    }
    return 0;
}
```

## Bundling with Your Application

1. Build libzpl for your target platforms
2. Ship the appropriate `.so`/`.dylib`/`.dll` alongside your binary
3. Set library search path at runtime:
   - **Linux**: `LD_LIBRARY_PATH` or rpath
   - **macOS**: `DYLD_LIBRARY_PATH` or rpath (`@executable_path`)
   - **Windows**: Same directory as .exe or in PATH

### Library Size

| Platform | Size |
|----------|------|
| Linux amd64 | ~20 MB |
| macOS universal | ~40 MB |
| Windows amd64 | ~20 MB |

The library is self-contained with no external dependencies.
