use std::env;
use std::fs::{self, File};
use std::io::{self, Read, Write};
use std::path::{Path, PathBuf};
use std::process::Command;

const GITHUB_RELEASE_URL: &str = "https://github.com/StirlingMarketingGroup/go-zpl/releases/download";

fn main() {
    let out_dir = PathBuf::from(env::var("OUT_DIR").unwrap());
    let target = env::var("TARGET").unwrap();

    // Get version from env var or use default
    let version = env::var("LIBZPL_VERSION").unwrap_or_else(|_| get_latest_version());

    // Determine the library file based on target
    let (archive_name, lib_name, extract_name) = get_library_info(&target);

    let lib_dir = out_dir.join("lib");
    fs::create_dir_all(&lib_dir).unwrap();

    let lib_path = lib_dir.join(lib_name);

    // Download and extract if the library doesn't exist
    if !lib_path.exists() {
        println!("cargo:warning=Downloading libzpl {} for {}", version, target);
        download_and_extract(&version, &archive_name, &lib_dir, extract_name)
            .expect("Failed to download libzpl");

        // Fix install name on macOS
        if target.contains("darwin") {
            fix_macos_install_name(&lib_path);
        }
    }

    // Tell cargo where to find the library
    println!("cargo:rustc-link-search=native={}", lib_dir.display());
    println!("cargo:rustc-link-lib=dylib=zpl");

    // On macOS, set rpath
    if target.contains("darwin") {
        println!("cargo:rustc-link-arg=-Wl,-rpath,@executable_path");
        println!("cargo:rustc-link-arg=-Wl,-rpath,{}", lib_dir.display());
    }

    // On Linux, set rpath
    if target.contains("linux") {
        println!("cargo:rustc-link-arg=-Wl,-rpath,$ORIGIN");
        println!("cargo:rustc-link-arg=-Wl,-rpath,{}", lib_dir.display());
    }

    // Rerun if these change
    println!("cargo:rerun-if-changed=build.rs");
    println!("cargo:rerun-if-env-changed=LIBZPL_VERSION");
    println!("cargo:rerun-if-env-changed=LIBZPL_PATH");
}

/// Get the latest release version from GitHub API
fn get_latest_version() -> String {
    // Try to get from GitHub API
    if let Ok(response) = reqwest::blocking::Client::new()
        .get("https://api.github.com/repos/StirlingMarketingGroup/go-zpl/releases/latest")
        .header("User-Agent", "zpl-rs-build")
        .send()
    {
        if response.status().is_success() {
            if let Ok(text) = response.text() {
                // Simple JSON parsing for tag_name
                if let Some(start) = text.find("\"tag_name\":\"") {
                    let rest = &text[start + 12..];
                    if let Some(end) = rest.find('"') {
                        let tag = &rest[..end];
                        // Strip 'v' prefix if present
                        return tag.trim_start_matches('v').to_string();
                    }
                }
            }
        }
    }

    // Fallback to known version
    "0.1.1".to_string()
}

fn get_library_info(target: &str) -> (&'static str, &'static str, &'static str) {
    match target {
        t if t.contains("x86_64") && t.contains("linux") => {
            ("libzpl-linux-amd64.tar.gz", "libzpl.so", "libzpl-linux-amd64.so")
        }
        t if t.contains("aarch64") && t.contains("linux") => {
            ("libzpl-linux-arm64.tar.gz", "libzpl.so", "libzpl-linux-arm64.so")
        }
        t if t.contains("darwin") => {
            // macOS universal binary works for both arm64 and x86_64
            ("libzpl-darwin.tar.gz", "libzpl.dylib", "libzpl-darwin-universal.dylib")
        }
        t if t.contains("x86_64") && t.contains("windows") => {
            ("libzpl-windows-amd64.zip", "zpl.dll", "libzpl-windows-amd64.dll")
        }
        t if t.contains("aarch64") && t.contains("windows") => {
            ("libzpl-windows-arm64.zip", "zpl.dll", "libzpl-windows-arm64.dll")
        }
        _ => panic!("Unsupported target: {}", target),
    }
}

fn download_and_extract(version: &str, archive_name: &str, lib_dir: &Path, extract_name: &str) -> io::Result<()> {
    let url = format!("{}/v{}/{}", GITHUB_RELEASE_URL, version, archive_name);

    // Download the archive
    let response = reqwest::blocking::get(&url)
        .map_err(|e| io::Error::new(io::ErrorKind::Other, e))?;

    if !response.status().is_success() {
        return Err(io::Error::new(
            io::ErrorKind::NotFound,
            format!("Failed to download {}: {}", url, response.status()),
        ));
    }

    let bytes = response.bytes()
        .map_err(|e| io::Error::new(io::ErrorKind::Other, e))?;

    // Extract based on archive type
    if archive_name.ends_with(".tar.gz") {
        extract_tar_gz(&bytes, lib_dir, extract_name)?;
    } else if archive_name.ends_with(".zip") {
        extract_zip(&bytes, lib_dir, extract_name)?;
    }

    Ok(())
}

fn extract_tar_gz(data: &[u8], lib_dir: &Path, extract_name: &str) -> io::Result<()> {
    use flate2::read::GzDecoder;
    use tar::Archive;

    let decoder = GzDecoder::new(data);
    let mut archive = Archive::new(decoder);

    for entry in archive.entries()? {
        let mut entry = entry?;
        let path = entry.path()?;

        if path.file_name().map(|n| n.to_str()) == Some(Some(extract_name)) {
            // Determine the output name (normalize to libzpl.so or libzpl.dylib)
            let out_name = if extract_name.contains("darwin") {
                "libzpl.dylib"
            } else {
                "libzpl.so"
            };
            let out_path = lib_dir.join(out_name);
            let mut out_file = File::create(&out_path)?;
            io::copy(&mut entry, &mut out_file)?;

            // Make executable on Unix
            #[cfg(unix)]
            {
                use std::os::unix::fs::PermissionsExt;
                fs::set_permissions(&out_path, fs::Permissions::from_mode(0o755))?;
            }

            return Ok(());
        }
    }

    Err(io::Error::new(
        io::ErrorKind::NotFound,
        format!("Library {} not found in archive", extract_name),
    ))
}

fn extract_zip(data: &[u8], lib_dir: &Path, extract_name: &str) -> io::Result<()> {
    use std::io::Cursor;
    use zip::ZipArchive;

    let cursor = Cursor::new(data);
    let mut archive = ZipArchive::new(cursor)
        .map_err(|e| io::Error::new(io::ErrorKind::Other, e))?;

    for i in 0..archive.len() {
        let mut file = archive.by_index(i)
            .map_err(|e| io::Error::new(io::ErrorKind::Other, e))?;

        if file.name().ends_with(extract_name) || file.name() == extract_name {
            let out_path = lib_dir.join("zpl.dll");
            let mut out_file = File::create(&out_path)?;

            let mut contents = Vec::new();
            file.read_to_end(&mut contents)?;
            out_file.write_all(&contents)?;

            return Ok(());
        }
    }

    Err(io::Error::new(
        io::ErrorKind::NotFound,
        format!("Library {} not found in archive", extract_name),
    ))
}

/// Fix the install name on macOS dylibs so they can be found at runtime
fn fix_macos_install_name(lib_path: &Path) {
    // Use install_name_tool to change the dylib's internal name
    let status = Command::new("install_name_tool")
        .args([
            "-id",
            "@rpath/libzpl.dylib",
            lib_path.to_str().unwrap(),
        ])
        .status();

    if let Err(e) = status {
        println!("cargo:warning=Failed to fix install name: {}", e);
    }
}
