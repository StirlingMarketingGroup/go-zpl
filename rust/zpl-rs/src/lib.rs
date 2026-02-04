//! # zpl-rs
//!
//! A Rust library for rendering ZPL (Zebra Programming Language) labels to images.
//!
//! This crate provides bindings to [go-zpl](https://github.com/StirlingMarketingGroup/go-zpl),
//! a native ZPL parser and renderer. Everything runs locally - no external services required.
//!
//! ## Thread Safety
//!
//! **Important:** The underlying Go library is not thread-safe for concurrent renders.
//! If you need to render from multiple threads, use a mutex or serialize your render calls.
//!
//! ## Quick Start
//!
//! ```rust
//! use zpl_rs::render;
//!
//! let zpl = r#"^XA
//! ^FO50,50^A0N,30,30^FDHello, World!^FS
//! ^FO50,100^BQN,2,5^FDMA,https://example.com^FS
//! ^XZ"#;
//!
//! // Render to PNG bytes
//! let png_bytes = render(zpl).expect("Failed to render ZPL");
//!
//! // Save to file
//! std::fs::write("label.png", &png_bytes).unwrap();
//! ```
//!
//! ## Advanced Usage
//!
//! ```rust
//! use zpl_rs::{render_with_options, RenderOptions, Dpi};
//!
//! let zpl = "^XA^FO50,50^A0N,30,30^FDHello!^FS^XZ";
//!
//! let options = RenderOptions {
//!     dpi: Dpi::Dpi300,
//!     width: Some(812),
//!     height: Some(1218),
//! };
//!
//! let png_bytes = render_with_options(zpl, &options).expect("Failed to render");
//! ```

use std::ffi::c_char;
use std::ffi::c_int;
use std::slice;

// FFI declarations
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

/// Error type for ZPL rendering operations.
#[derive(Debug, Clone, PartialEq, Eq)]
pub enum Error {
    /// Failed to parse the ZPL content.
    ParseError,
    /// Failed to render the label to an image.
    RenderError,
    /// Internal library error.
    InternalError,
    /// Unknown error with the given code.
    Unknown(i32),
}

impl std::fmt::Display for Error {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        match self {
            Error::ParseError => write!(f, "Failed to parse ZPL"),
            Error::RenderError => write!(f, "Failed to render label"),
            Error::InternalError => write!(f, "Internal library error"),
            Error::Unknown(code) => write!(f, "Unknown error (code: {})", code),
        }
    }
}

impl std::error::Error for Error {}

impl From<c_int> for Error {
    fn from(code: c_int) -> Self {
        match code {
            -1 => Error::ParseError,
            -2 => Error::RenderError,
            -3 => Error::InternalError,
            n => Error::Unknown(n),
        }
    }
}

/// Result type for ZPL operations.
pub type Result<T> = std::result::Result<T, Error>;

/// Printer DPI settings.
#[derive(Debug, Clone, Copy, PartialEq, Eq, Default)]
pub enum Dpi {
    /// 203 DPI (default for most Zebra printers)
    #[default]
    Dpi203,
    /// 300 DPI
    Dpi300,
    /// 600 DPI
    Dpi600,
}

impl Dpi {
    fn as_c_int(self) -> c_int {
        match self {
            Dpi::Dpi203 => 203,
            Dpi::Dpi300 => 300,
            Dpi::Dpi600 => 600,
        }
    }
}

/// Options for rendering ZPL labels.
#[derive(Debug, Clone, Default)]
pub struct RenderOptions {
    /// Printer DPI (default: 203)
    pub dpi: Dpi,
    /// Label width in dots (None = auto-detect from ZPL)
    pub width: Option<i32>,
    /// Label height in dots (None = auto-detect from ZPL)
    pub height: Option<i32>,
}

impl RenderOptions {
    /// Create new render options with default settings.
    pub fn new() -> Self {
        Self::default()
    }

    /// Set the DPI.
    pub fn dpi(mut self, dpi: Dpi) -> Self {
        self.dpi = dpi;
        self
    }

    /// Set the label width in dots.
    pub fn width(mut self, width: i32) -> Self {
        self.width = Some(width);
        self
    }

    /// Set the label height in dots.
    pub fn height(mut self, height: i32) -> Self {
        self.height = Some(height);
        self
    }

    /// Set the label size in dots.
    pub fn size(mut self, width: i32, height: i32) -> Self {
        self.width = Some(width);
        self.height = Some(height);
        self
    }
}

/// Render ZPL content to PNG image data using default settings (203 DPI, auto dimensions).
///
/// # Arguments
///
/// * `zpl` - The ZPL content to render
///
/// # Returns
///
/// PNG image data as a `Vec<u8>`
///
/// # Example
///
/// ```rust
/// use zpl_rs::render;
///
/// let zpl = "^XA^FO50,50^A0N,30,30^FDHello!^FS^XZ";
/// let png = render(zpl).expect("render failed");
///
/// // Verify it's a valid PNG
/// assert_eq!(&png[0..4], &[0x89, b'P', b'N', b'G']);
/// ```
pub fn render(zpl: &str) -> Result<Vec<u8>> {
    render_bytes(zpl.as_bytes())
}

/// Render ZPL content (as bytes) to PNG image data using default settings.
///
/// This is useful when working with ZPL data that may not be valid UTF-8.
pub fn render_bytes(zpl: &[u8]) -> Result<Vec<u8>> {
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
        return Err(Error::from(result));
    }

    // Copy to Rust-owned Vec and free the C memory
    let png_data = unsafe {
        let slice = slice::from_raw_parts(png_ptr as *const u8, png_len as usize);
        let owned = slice.to_vec();
        zpl_free(png_ptr);
        owned
    };

    Ok(png_data)
}

/// Render ZPL content to PNG image data with custom options.
///
/// # Arguments
///
/// * `zpl` - The ZPL content to render
/// * `options` - Rendering options (DPI, dimensions)
///
/// # Example
///
/// ```rust
/// use zpl_rs::{render_with_options, RenderOptions, Dpi};
///
/// let zpl = "^XA^FO50,50^A0N,30,30^FDHello!^FS^XZ";
/// let options = RenderOptions::new()
///     .dpi(Dpi::Dpi300)
///     .size(812, 1218);
///
/// let png = render_with_options(zpl, &options).expect("render failed");
/// ```
pub fn render_with_options(zpl: &str, options: &RenderOptions) -> Result<Vec<u8>> {
    render_bytes_with_options(zpl.as_bytes(), options)
}

/// Render ZPL content (as bytes) to PNG image data with custom options.
pub fn render_bytes_with_options(zpl: &[u8], options: &RenderOptions) -> Result<Vec<u8>> {
    let mut png_ptr: *mut c_char = std::ptr::null_mut();
    let mut png_len: c_int = 0;

    let result = unsafe {
        zpl_render_png(
            zpl.as_ptr() as *const c_char,
            zpl.len() as c_int,
            options.dpi.as_c_int(),
            options.width.unwrap_or(0),
            options.height.unwrap_or(0),
            &mut png_ptr,
            &mut png_len,
        )
    };

    if result != 0 {
        return Err(Error::from(result));
    }

    let png_data = unsafe {
        let slice = slice::from_raw_parts(png_ptr as *const u8, png_len as usize);
        let owned = slice.to_vec();
        zpl_free(png_ptr);
        owned
    };

    Ok(png_data)
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_render_simple() {
        let zpl = "^XA^FO50,50^A0N,30,30^FDHello Rust!^FS^XZ";
        let png = render(zpl).expect("render failed");

        // Check PNG magic bytes
        assert!(png.len() > 8);
        assert_eq!(&png[0..4], &[0x89, b'P', b'N', b'G']);
    }

    #[test]
    fn test_render_with_options() {
        let zpl = "^XA^FO50,50^A0N,30,30^FDHello!^FS^XZ";
        let options = RenderOptions::new()
            .dpi(Dpi::Dpi300)
            .size(812, 1218);

        let png = render_with_options(zpl, &options).expect("render failed");
        assert!(png.len() > 8);
        assert_eq!(&png[0..4], &[0x89, b'P', b'N', b'G']);
    }

    #[test]
    fn test_render_with_barcode() {
        let zpl = r#"^XA
^FO50,50^A0N,30,30^FDTest Label^FS
^FO50,100^BCN,100,Y,N,N,N^FD123456789^FS
^FO50,250^BQN,2,5^FDMA,https://example.com^FS
^XZ"#;

        let png = render(zpl).expect("render failed");
        assert!(png.len() > 1000); // Should be a decent size with barcodes
    }

    #[test]
    fn test_empty_label() {
        // The parser is lenient - it will render even minimal/empty ZPL
        // This just produces a blank label, not an error
        let result = render("^XA^XZ");
        assert!(result.is_ok());
    }

    #[test]
    fn test_render_bytes() {
        let zpl = b"^XA^FO50,50^A0N,30,30^FDBytes!^FS^XZ";
        let png = render_bytes(zpl).expect("render failed");
        assert_eq!(&png[0..4], &[0x89, b'P', b'N', b'G']);
    }
}
