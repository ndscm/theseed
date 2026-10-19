//! http(s) transport: the optional cache server and credential helper.
//!
//! This layer knows nothing about the on-disk cache layout or entry data model;
//! it just moves bytes between URLs and caller-provided paths, applying the
//! credential helper.

use ::anyhow::Context;
use ::anyhow::Result;
use ::anyhow::bail;
use ::serde::Deserialize;
use ::std::collections::BTreeMap;
use ::std::fs::File;
use ::std::path::Path;
use ::std::path::PathBuf;
use ::std::process::Command;

/// Environment variable holding the optional cache server base URL.
pub const XSLASH_CACHE_SERVER: &str = "XSLASH_CACHE_SERVER";
/// Environment variable holding the optional credential helper path.
pub const XSLASH_CREDENTIAL_HELPER: &str = "XSLASH_CREDENTIAL_HELPER";

/// An http(s) client wired to the configured cache server and credential
/// helper.
pub struct Http {
    /// Suppress xslash's messages.
    silent: bool,

    agent: ureq::Agent,
    cache_server: Option<String>,
    credential_helper: Option<PathBuf>,
}

/// The JSON a credential helper prints on stdout (same shape as bazel's).
#[derive(Deserialize)]
struct Credentials {
    #[serde(default)]
    headers: BTreeMap<String, Vec<String>>,
}

impl Http {
    /// Build a client from the environment.
    pub fn from_env(silent: bool) -> Http {
        let cache_server = std::env::var(XSLASH_CACHE_SERVER)
            .ok()
            .filter(|s| !s.is_empty())
            .map(|s| s.trim_end_matches('/').to_string());
        let credential_helper = std::env::var_os(XSLASH_CREDENTIAL_HELPER)
            .filter(|s| !s.is_empty())
            .map(PathBuf::from);
        Http {
            silent,
            agent: ureq::Agent::new_with_defaults(),
            cache_server,
            credential_helper,
        }
    }

    /// Ask the credential helper for headers to attach to `method url`. Returns
    /// an empty list when no helper is configured or it prints nothing. An
    /// error means the request should be skipped (helper exited non-zero).
    fn credentials(&self, method: &str, url: &str) -> Result<Vec<(String, String)>> {
        let Some(helper) = &self.credential_helper else {
            return Ok(Vec::new());
        };
        let output = Command::new(helper)
            .arg(method)
            .arg(url)
            .output()
            .with_context(|| format!("running credential helper {}", helper.display()))?;
        if !output.status.success() {
            bail!("credential helper exited non-zero for {method} {url}");
        }
        if output.stdout.iter().all(u8::is_ascii_whitespace) {
            return Ok(Vec::new());
        }
        let creds: Credentials =
            serde_json::from_slice(&output.stdout).context("parsing credential helper output")?;
        Ok(creds
            .headers
            .into_iter()
            .map(|(name, values)| (name, values.join(", ")))
            .collect())
    }

    /// GET `url`, streaming the body into `dest`.
    pub fn get(&self, url: &str, dest: &Path) -> Result<()> {
        if !self.silent {
            eprintln!("xslash: get {url}");
        }
        let headers = self.credentials("GET", url)?;
        let mut req = self.agent.get(url);
        for (name, value) in &headers {
            req = req.header(name, value);
        }
        let response = req.call().with_context(|| format!("GET {url}"))?;
        let mut reader = response.into_body().into_reader();
        let mut file =
            File::create(dest).with_context(|| format!("creating {}", dest.display()))?;
        ::std::io::copy(&mut reader, &mut file)
            .with_context(|| format!("writing {}", dest.display()))?;
        Ok(())
    }

    /// PUT the contents of `file` to `url`.
    fn put(&self, url: &str, file: &Path) -> Result<()> {
        if !self.silent {
            eprintln!("xslash: put {url}");
        }
        let headers = self.credentials("PUT", url)?;
        let bytes = std::fs::read(file).with_context(|| format!("reading {}", file.display()))?;
        let mut req = self.agent.put(url);
        for (name, value) in &headers {
            req = req.header(name, value);
        }
        req.send(&bytes).with_context(|| format!("PUT {url}"))?;
        Ok(())
    }

    /// GET the artifact `<hash>/<digest>` from the cache server into `dest`.
    /// Returns whether a cache server is configured and the GET succeeded. The
    /// result is *not* verified; the caller must check its digest.
    pub fn cache_server_get(&self, hash: &str, digest: &str, dest: &Path) -> bool {
        let Some(server) = &self.cache_server else {
            return false;
        };
        self.get(&format!("{server}/{hash}/{digest}"), dest).is_ok()
    }

    /// Best-effort PUT of `file` back to the cache server as `<hash>/<digest>`.
    /// A no-op when no cache server is configured; failures are ignored.
    pub fn cache_server_put(&self, hash: &str, digest: &str, file: &Path) {
        if let Some(server) = &self.cache_server {
            let _ = self.put(&format!("{server}/{hash}/{digest}"), file);
        }
    }
}
