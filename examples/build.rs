// Copyright (c) 2022, Mysten Labs, Inc.
// SPDX-License-Identifier: Apache-2.0


fn main() -> Result<(), Box<dyn std::error::Error>> {
    // println!("cargo:rerun-if-changed=protos/narwhal.proto");
    println!("cargo");
    tonic_build::compile_protos("protos/narwhal.proto")?;
    Ok(())
}
