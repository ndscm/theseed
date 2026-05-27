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
    #[arg(long, default_value = "skeleton.dotslash.json")]
    skeleton: String,

    #[arg(long, default_value = "v0.0.0")]
    tag: String,

    #[arg(long, default_value_t = false)]
    no_format: bool,
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

    let skeleton_bytes =
        std::fs::read_to_string(&args.skeleton).context("reading skeleton file")?;
    let mut result: DotSlash = serde_json::from_str(&skeleton_bytes).context("parsing skeleton")?;

    for (platform_name, platform) in &mut result.platforms {
        for provider in &mut platform.providers {
            provider.url = provider.url.replace("{{TAG}}", &args.tag);

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
            if args.no_format {
                platform.format = None;
            } else {
                platform.format = entry.format;
            }
        }
        eprintln!("Fetched: platform={:#?}", platform);
    }

    let mut stdout = std::io::stdout().lock();
    stdout.write_all(b"#!/usr/bin/env dotslash\n\n")?;
    serde_json::to_writer_pretty(&mut stdout, &result)?;
    stdout.write_all(b"\n")?;
    Ok(())
}

fn main() {
    if let Err(e) = run() {
        eprintln!("{e:#}");
        std::process::exit(1);
    }
}
