use std::env;
use std::fs::{self, File};
use std::io::{self, Read, Write};
use std::path::{Path, PathBuf};
use std::process::Command;

const GITHUB_RELEASE_URL: &str =
    "https://github.com/StirlingMarketingGroup/go-zpl/releases/download";

/// Where the library comes from. Chosen once, before anything touches the network.
enum Source {
    /// A library the user already has on disk (LIBZPL_PATH).
    File(PathBuf),
    /// A GitHub release to download (LIBZPL_VERSION or the latest release).
    Release(String),
}

/// Everything the build script decides per target, decided in one place.
struct Platform {
    archive_name: &'static str,
    extract_name: &'static str,
    lib_name: &'static str,
    /// Windows links through `kind = "raw-dylib"` in src/lib.rs instead: the release
    /// archives ship only zpl.dll, and `dylib=zpl` would make MSVC demand an import library.
    link_lib: bool,
    /// Loader-relative rpaths for this crate's own tests and examples. They never reach
    /// dependents (Cargo scopes rustc-link-arg to the emitting package), which is why the
    /// metadata in main() and the README's bundling section exist.
    rpaths: &'static [&'static str],
    /// The prebuilt dylib carries the install name it was built with; rewriting it to
    /// @rpath is what lets the rpaths above (and a consumer's) find it.
    rewrite_install_name: bool,
}

fn main() {
    println!("cargo:rerun-if-changed=build.rs");
    println!("cargo:rerun-if-env-changed=LIBZPL_VERSION");
    println!("cargo:rerun-if-env-changed=LIBZPL_PATH");
    println!("cargo:rerun-if-env-changed=LIBZPL_COPY_TO");

    let out_dir = PathBuf::from(env::var("OUT_DIR").unwrap());
    let target = env::var("TARGET").unwrap();
    let platform = platform(&target);

    let lib_dir = out_dir.join("lib");
    fs::create_dir_all(&lib_dir).unwrap();
    let lib_path = lib_dir.join(platform.lib_name);
    let stamp_path = lib_dir.join(".source");

    let source = select_source();
    stage(&source, &platform, &target, &lib_path, &stamp_path);

    println!("cargo:rustc-link-search=native={}", lib_dir.display());
    if platform.link_lib {
        println!("cargo:rustc-link-lib=dylib=zpl");
    }
    for rpath in platform.rpaths {
        println!("cargo:rustc-link-arg=-Wl,-rpath,{}", rpath);
    }
    if !platform.rpaths.is_empty() {
        println!("cargo:rustc-link-arg=-Wl,-rpath,{}", lib_dir.display());
    }

    let version = match &source {
        Source::Release(v) => v.as_str(),
        Source::File(_) => "file",
    };
    println!("cargo:lib_dir={}", lib_dir.display());
    println!("cargo:lib_path={}", lib_path.display());
    println!("cargo:lib_name={}", platform.lib_name);
    println!("cargo:version={}", version);

    copy_to(&source, &platform, &lib_path);
}

fn platform(target: &str) -> Platform {
    let unix_rpaths: &'static [&'static str] = &["$ORIGIN"];
    let mac_rpaths: &'static [&'static str] = &["@executable_path"];
    match target {
        t if t.contains("x86_64") && t.contains("linux") => Platform {
            archive_name: "libzpl-linux-amd64.tar.gz",
            extract_name: "libzpl-linux-amd64.so",
            lib_name: "libzpl.so",
            link_lib: true,
            rpaths: unix_rpaths,
            rewrite_install_name: false,
        },
        t if t.contains("aarch64") && t.contains("linux") => Platform {
            archive_name: "libzpl-linux-arm64.tar.gz",
            extract_name: "libzpl-linux-arm64.so",
            lib_name: "libzpl.so",
            link_lib: true,
            rpaths: unix_rpaths,
            rewrite_install_name: false,
        },
        // One universal binary serves both macOS architectures.
        t if t.contains("darwin") => Platform {
            archive_name: "libzpl-darwin.tar.gz",
            extract_name: "libzpl-darwin-universal.dylib",
            lib_name: "libzpl.dylib",
            link_lib: true,
            rpaths: mac_rpaths,
            rewrite_install_name: true,
        },
        t if t.contains("x86_64") && t.contains("windows") => Platform {
            archive_name: "libzpl-windows-amd64.zip",
            extract_name: "libzpl-windows-amd64.dll",
            lib_name: "zpl.dll",
            link_lib: false,
            rpaths: &[],
            rewrite_install_name: false,
        },
        t if t.contains("aarch64") && t.contains("windows") => Platform {
            archive_name: "libzpl-windows-arm64.zip",
            extract_name: "libzpl-windows-arm64.dll",
            lib_name: "zpl.dll",
            link_lib: false,
            rpaths: &[],
            rewrite_install_name: false,
        },
        _ => panic!("Unsupported target: {}", target),
    }
}

fn select_source() -> Source {
    match env::var("LIBZPL_PATH") {
        Ok(raw) if !raw.is_empty() => {
            let path = PathBuf::from(&raw);
            if !path.is_file() {
                panic!(
                    "LIBZPL_PATH is set to {} but that file does not exist",
                    path.display()
                );
            }
            // Keep the path the user set (absolute, but not canonicalized) so the stamp
            // matches LIBZPL_PATH on platforms where /tmp is a symlink.
            let path = if path.is_absolute() {
                path
            } else {
                env::current_dir().expect("current directory").join(path)
            };
            Source::File(path)
        }
        _ => {
            let version = env::var("LIBZPL_VERSION").unwrap_or_else(|_| get_latest_version());
            Source::Release(version)
        }
    }
}

/// Puts the library at `lib_path` with `stamp_path` recording where it came from.
/// The library is built up as a `.partial` sibling and renamed into place only once it is
/// complete, and the stamp is removed first and rewritten last, so an interrupted run can
/// never leave a half-written file that a later run mistakes for a finished one.
fn stage(source: &Source, platform: &Platform, target: &str, lib_path: &Path, stamp_path: &Path) {
    let stamp = match source {
        Source::Release(version) => format!("release {}", version),
        Source::File(path) => {
            println!("cargo:rerun-if-changed={}", path.display());
            format!("file {}", path.display())
        }
    };
    // A user-supplied file is always re-copied: it is cheap, and the rerun rules already
    // limit how often this runs. A release is reused while the stamp still matches it.
    let reuse = matches!(source, Source::Release(_))
        && lib_path.exists()
        && fs::read_to_string(stamp_path).ok().as_deref() == Some(stamp.as_str());
    if reuse {
        return;
    }

    if let Err(e) = fs::remove_file(stamp_path) {
        if e.kind() != io::ErrorKind::NotFound {
            panic!("Failed to remove {}: {}", stamp_path.display(), e);
        }
    }

    let partial = lib_path.with_extension("partial");
    match source {
        Source::Release(version) => {
            println!(
                "cargo:warning=Downloading libzpl {} for {}",
                version, target
            );
            download_and_extract(version, platform, &partial).expect("Failed to download libzpl");
        }
        Source::File(path) => {
            fs::copy(path, &partial).unwrap_or_else(|e| {
                panic!(
                    "Failed to copy LIBZPL_PATH {} to {}: {}",
                    path.display(),
                    partial.display(),
                    e
                );
            });
        }
    }
    if platform.rewrite_install_name {
        // Only ever the staged copy, never the user's original file.
        fix_macos_install_name(&partial);
    }
    fs::rename(&partial, lib_path).unwrap_or_else(|e| {
        panic!(
            "Failed to move {} to {}: {}",
            partial.display(),
            lib_path.display(),
            e
        );
    });
    fs::write(stamp_path, &stamp).expect("Failed to write libzpl source stamp");
}

/// The opt-in LIBZPL_COPY_TO hook. This deliberately writes outside OUT_DIR, which Cargo
/// discourages for build scripts, because it is the one way an app can stage the library
/// for a bundler without scanning target/.
fn copy_to(source: &Source, platform: &Platform, lib_path: &Path) {
    let copy_to = match env::var("LIBZPL_COPY_TO") {
        Ok(dir) if !dir.is_empty() => dir,
        _ => return,
    };
    let dest_dir = PathBuf::from(&copy_to);
    fs::create_dir_all(&dest_dir).unwrap_or_else(|e| {
        panic!(
            "LIBZPL_COPY_TO={}: failed to create directory {}: {}",
            copy_to,
            dest_dir.display(),
            e
        );
    });
    let dest = dest_dir.join(platform.lib_name);
    if let Source::File(original) = source {
        // The staged copy carries a rewritten install name; writing it back over the
        // user's file, whether the destination is that path, a symlink to it, or a hard
        // link of it, would silently alter and unsign the library they supplied.
        if dest.exists() && same_file(&dest, original) {
            panic!(
                "LIBZPL_COPY_TO={} would overwrite LIBZPL_PATH={}; choose a different directory",
                copy_to,
                original.display()
            );
        }
    }
    // Cargo takes the build script's reference time at invocation, so a file this script
    // writes and also watches reads as changed on the next build. Copying only when the
    // content differs lets it settle: a deleted or replaced copy is remade, an intact one
    // is left alone and the script goes fresh.
    let staged = fs::read(lib_path)
        .unwrap_or_else(|e| panic!("failed to read {}: {}", lib_path.display(), e));
    if fs::read(&dest).ok().as_deref() != Some(staged.as_slice()) {
        fs::copy(lib_path, &dest).unwrap_or_else(|e| {
            panic!(
                "LIBZPL_COPY_TO={}: failed to copy {} to {}: {}",
                copy_to,
                lib_path.display(),
                dest.display(),
                e
            );
        });
    }
    println!("cargo:rerun-if-changed={}", dest.display());
}

/// File identity (device + inode), following symlinks, so hard links count as the same file.
#[cfg(unix)]
fn same_file(a: &Path, b: &Path) -> bool {
    use std::os::unix::fs::MetadataExt;
    match (fs::metadata(a), fs::metadata(b)) {
        (Ok(a), Ok(b)) => a.dev() == b.dev() && a.ino() == b.ino(),
        _ => false,
    }
}

/// Windows exposes no stable file identity, so fall back to resolved paths (symlinks, not
/// hard links). Nothing rewrites the staged copy there, so the content check already skips
/// an identical destination and this only guards the path itself.
#[cfg(not(unix))]
fn same_file(a: &Path, b: &Path) -> bool {
    match (fs::canonicalize(a), fs::canonicalize(b)) {
        (Ok(a), Ok(b)) => a == b,
        _ => false,
    }
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

fn download_and_extract(version: &str, platform: &Platform, out_path: &Path) -> io::Result<()> {
    let url = format!(
        "{}/v{}/{}",
        GITHUB_RELEASE_URL, version, platform.archive_name
    );

    let response = reqwest::blocking::get(&url).map_err(io::Error::other)?;
    if !response.status().is_success() {
        return Err(io::Error::new(
            io::ErrorKind::NotFound,
            format!("Failed to download {}: {}", url, response.status()),
        ));
    }
    let bytes = response.bytes().map_err(io::Error::other)?;

    if platform.archive_name.ends_with(".tar.gz") {
        extract_tar_gz(&bytes, platform.extract_name, out_path)
    } else if platform.archive_name.ends_with(".zip") {
        extract_zip(&bytes, platform.extract_name, out_path)
    } else {
        Err(io::Error::new(
            io::ErrorKind::InvalidInput,
            format!("Unknown archive type: {}", platform.archive_name),
        ))
    }
}

fn extract_tar_gz(data: &[u8], extract_name: &str, out_path: &Path) -> io::Result<()> {
    use flate2::read::GzDecoder;
    use tar::Archive;

    let decoder = GzDecoder::new(data);
    let mut archive = Archive::new(decoder);

    for entry in archive.entries()? {
        let mut entry = entry?;
        let path = entry.path()?;

        if path.file_name().map(|n| n.to_str()) == Some(Some(extract_name)) {
            let mut out_file = File::create(out_path)?;
            io::copy(&mut entry, &mut out_file)?;

            #[cfg(unix)]
            {
                use std::os::unix::fs::PermissionsExt;
                fs::set_permissions(out_path, fs::Permissions::from_mode(0o755))?;
            }

            return Ok(());
        }
    }

    Err(io::Error::new(
        io::ErrorKind::NotFound,
        format!("Library {} not found in archive", extract_name),
    ))
}

fn extract_zip(data: &[u8], extract_name: &str, out_path: &Path) -> io::Result<()> {
    use std::io::Cursor;
    use zip::ZipArchive;

    let cursor = Cursor::new(data);
    let mut archive = ZipArchive::new(cursor).map_err(io::Error::other)?;

    for i in 0..archive.len() {
        let mut file = archive.by_index(i).map_err(io::Error::other)?;

        if file.name().ends_with(extract_name) || file.name() == extract_name {
            let mut out_file = File::create(out_path)?;
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

/// Rewrite the dylib's install name to @rpath so rpaths resolve it at runtime.
fn fix_macos_install_name(lib_path: &Path) {
    // A silently unmodified id would make a relocated binary look for the dylib by its
    // original install name instead of the rpath the README promises, so fail the build.
    let status = Command::new("install_name_tool")
        .args(["-id", "@rpath/libzpl.dylib", lib_path.to_str().unwrap()])
        .status()
        .unwrap_or_else(|e| {
            panic!(
                "failed to run install_name_tool on {}: {}",
                lib_path.display(),
                e
            )
        });
    if !status.success() {
        panic!(
            "install_name_tool -id @rpath/libzpl.dylib {} failed: {}",
            lib_path.display(),
            status
        );
    }
}
