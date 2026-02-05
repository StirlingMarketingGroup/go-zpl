use rxing::common::HybridBinarizer;
use rxing::maxicode::MaxiCodeReader;
use rxing::{BinaryBitmap, BufferedImageLuminanceSource, DecodeHints, Reader};
use std::env;
use std::io::Write;
use std::process;

fn main() {
    let args: Vec<String> = env::args().collect();
    if args.len() != 2 {
        eprintln!("Usage: maxicode-decode <image-path>");
        process::exit(1);
    }

    let path = &args[1];
    let img = match image::open(path) {
        Ok(img) => img,
        Err(e) => {
            eprintln!("Failed to open image: {}", e);
            process::exit(1);
        }
    };

    let lum = BufferedImageLuminanceSource::new(img);
    let mut bitmap = BinaryBitmap::new(HybridBinarizer::new(lum));

    let mut hints = DecodeHints::default();
    hints.PureBarcode = Some(true);
    hints.TryHarder = Some(true);

    let mut reader = MaxiCodeReader::default();
    match reader.decode_with_hints(&mut bitmap, &hints) {
        Ok(result) => {
            // Write raw bytes to stdout so control characters are preserved
            std::io::stdout()
                .write_all(result.getText().as_bytes())
                .unwrap();
        }
        Err(e) => {
            eprintln!("Decode failed: {}", e);
            process::exit(1);
        }
    }
}
