use std::process::Command;

fn main() {
    let output = Command::new("git")
        .args(&["describe", "--tags", "--always", "--dirty"])
        .output()
        .expect("Failed to execute git");

    if output.status.success() {
        let git_version = String::from_utf8(output.stdout)
            .expect("Invalid UTF-8")
            .trim()
            .to_string();
        println!("cargo:rustc-env=CARGO_PKG_VERSION={}", git_version);
    }

    println!("cargo:rerun-if-changed=.git/HEAD");
    println!("cargo:rerun-if-changed=.git/index");
}
