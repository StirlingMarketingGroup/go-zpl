// Example Rust FFI bindings for libzpl
//
// Add to Cargo.toml:
//   [build-dependencies]
//   cc = "1.0"
//
// Or link directly:
//   #[link(name = "zpl")]
//
// For marlin, you'd typically:
// 1. Bundle libzpl.so/dylib/dll with the app
// 2. Use build.rs to set the library search path

use std::ffi::c_char;
use std::ffi::c_int;
use std::slice;

#[link(name = "zpl")]
extern "C" {
    fn zpl_render_png(
        zpl_data: *const c_char,
        zpl_len: c_int,
        dpi: c_int,
        width: c_int,
        height: c_int,
        png_out: *mut *mut c_char,
        png_len: *mut c_int,
    ) -> c_int;

    fn zpl_render_png_simple(
        zpl_data: *const c_char,
        zpl_len: c_int,
        png_out: *mut *mut c_char,
        png_len: *mut c_int,
    ) -> c_int;

    fn zpl_free(ptr: *mut c_char);
}

/// Error codes from libzpl
#[derive(Debug, Clone, Copy, PartialEq, Eq)]
pub enum ZplError {
    ParseError,
    RenderError,
    InternalError,
    Unknown(i32),
}

impl From<c_int> for ZplError {
    fn from(code: c_int) -> Self {
        match code {
            -1 => ZplError::ParseError,
            -2 => ZplError::RenderError,
            -3 => ZplError::InternalError,
            n => ZplError::Unknown(n),
        }
    }
}

/// Render ZPL to PNG bytes
///
/// # Arguments
/// * `zpl` - The ZPL content as bytes
/// * `dpi` - Printer DPI (203, 300, or 600)
/// * `width` - Label width in dots (0 for auto)
/// * `height` - Label height in dots (0 for auto)
///
/// # Returns
/// PNG image data as a Vec<u8>
pub fn render_zpl(zpl: &[u8], dpi: i32, width: i32, height: i32) -> Result<Vec<u8>, ZplError> {
    let mut png_ptr: *mut c_char = std::ptr::null_mut();
    let mut png_len: c_int = 0;

    let result = unsafe {
        zpl_render_png(
            zpl.as_ptr() as *const c_char,
            zpl.len() as c_int,
            dpi,
            width,
            height,
            &mut png_ptr,
            &mut png_len,
        )
    };

    if result != 0 {
        return Err(ZplError::from(result));
    }

    // Copy the data to a Rust-owned Vec
    let png_data = unsafe {
        let slice = slice::from_raw_parts(png_ptr as *const u8, png_len as usize);
        let owned = slice.to_vec();
        zpl_free(png_ptr); // Free the C-allocated memory
        owned
    };

    Ok(png_data)
}

/// Render ZPL to PNG with default settings (203 DPI, auto dimensions)
pub fn render_zpl_simple(zpl: &[u8]) -> Result<Vec<u8>, ZplError> {
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
        return Err(ZplError::from(result));
    }

    let png_data = unsafe {
        let slice = slice::from_raw_parts(png_ptr as *const u8, png_len as usize);
        let owned = slice.to_vec();
        zpl_free(png_ptr);
        owned
    };

    Ok(png_data)
}

// Example usage for marlin thumbnail generation:
//
// fn generate_zpl_thumbnail(file_path: &Path) -> Result<Vec<u8>, Box<dyn Error>> {
//     let zpl_content = std::fs::read(file_path)?;
//     let png = render_zpl_simple(&zpl_content)?;
//     Ok(png)
// }

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_render_simple() {
        let zpl = b"^XA^FO50,50^A0N,30,30^FDHello Rust!^FS^XZ";
        let png = render_zpl_simple(zpl).expect("render failed");

        // Check PNG magic bytes
        assert!(png.len() > 8);
        assert_eq!(&png[0..4], &[0x89, b'P', b'N', b'G']);
    }
}
