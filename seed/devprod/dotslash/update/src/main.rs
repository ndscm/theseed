use anyhow::Context;
use anyhow::Result;
use clap::Parser;
use serde::Deserialize;
use serde::Serialize;
use std::collections::BTreeMap;
use std::io::Write;
use std::process::Command;

#[derive(Parser)]
struct Args {
    /// Skeleton file(s) to process. May be passed multiple times, but multiple
    /// values are only allowed when `--outdir` is set.
    #[arg(long, default_value = "skeleton.dotslash.json")]
    skeleton: Vec<String>,

    /// Placeholder replacement in `KEY=VALUE` form; each occurrence of
    /// `{{KEY}}` in paths and provider URLs is replaced with `VALUE`.
    /// May be passed multiple times.
    #[arg(long = "replace", value_name = "KEY=VALUE")]
    replacements: Vec<String>,

    /// If set, write the result to `<outdir>/<basename>.dotslash` instead of
    /// stdout. The basename is derived from the skeleton file name by stripping
    /// a trailing `.dotslash.json` suffix, falling back to the whole file name.
    #[arg(long)]
    outdir: Option<String>,
}

#[derive(Debug, Serialize, Deserialize)]
struct Provider {
    url: String,
}

#[derive(Debug, Serialize, Deserialize)]
struct Platform {
    path: String,

    #[serde(skip_serializing_if = "Option::is_none")]
    size: Option<u64>,

    #[serde(skip_serializing_if = "Option::is_none")]
    hash: Option<String>,

    #[serde(skip_serializing_if = "Option::is_none")]
    digest: Option<String>,

    #[serde(skip_serializing_if = "Option::is_none")]
    format: Option<String>,

    providers: Vec<Provider>,
}

#[derive(Serialize, Deserialize)]
struct DotSlash {
    name: String,
    platforms: BTreeMap<String, Platform>,
}

fn run() -> Result<()> {
    let args = Args::parse();

    let mut replacements: Vec<(String, String)> = Vec::new();
    for replacement in &args.replacements {
        let (key, value) = replacement.split_once('=').context(format!(
            "invalid --replace, expected KEY=VALUE: {replacement}"
        ))?;
        replacements.push((format!("{{{{{key}}}}}"), value.to_string()));
    }
    let apply = |mut s: String| {
        for (placeholder, value) in &replacements {
            s = s.replace(placeholder.as_str(), value);
        }
        s
    };

    if args.skeleton.len() > 1 && args.outdir.is_none() {
        anyhow::bail!("multiple --skeleton values require --outdir");
    }

    for skeleton in &args.skeleton {
        let skeleton_bytes = std::fs::read_to_string(skeleton).context("reading skeleton file")?;
        let mut result: DotSlash =
            serde_json::from_str(&skeleton_bytes).context("parsing skeleton")?;

        for (platform_name, platform) in &mut result.platforms {
            platform.path = apply(std::mem::take(&mut platform.path));
            for provider in &mut platform.providers {
                provider.url = apply(std::mem::take(&mut provider.url));

                eprintln!("Fetching: platform={} url={}", platform_name, provider.url);
                let output = Command::new("dotslash")
                    .args(["--", "create-url-entry", &provider.url])
                    .stderr(std::process::Stdio::inherit())
                    .output()
                    .context("running dotslash")?;
                if !output.status.success() {
                    anyhow::bail!("dotslash create-url-entry failed for {}", provider.url);
                }

                let entry: Platform = serde_json::from_slice(&output.stdout)
                    .context(format!("parsing dotslash output for {}", provider.url))?;

                platform.size = entry.size;
                platform.hash = entry.hash;
                platform.digest = entry.digest;
            }
            eprintln!("Fetched: platform={:#?}", platform);
        }

        let mut out: Vec<u8> = Vec::new();
        out.extend_from_slice(b"#!/usr/bin/env dotslash\n\n");
        serde_json::to_writer_pretty(&mut out, &result)?;
        out.push(b'\n');

        if let Some(outdir) = &args.outdir {
            let file_name = std::path::Path::new(skeleton)
                .file_name()
                .and_then(|n| n.to_str())
                .context("skeleton path has no file name")?;
            let basename = file_name
                .strip_suffix(".dotslash.json")
                .unwrap_or(file_name);
            let path = std::path::Path::new(outdir).join(format!("{basename}.dotslash"));
            std::fs::write(&path, &out).context(format!("writing {}", path.display()))?;
            eprintln!("Wrote: {}", path.display());
        } else {
            let mut stdout = std::io::stdout().lock();
            stdout.write_all(&out)?;
        }
    }
    Ok(())
}

fn main() {
    if let Err(e) = run() {
        eprintln!("{e:#}");
        std::process::exit(1);
    }
}
