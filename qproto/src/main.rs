fn main() -> Result<(), Box<dyn std::error::Error>> {
    // println!("cargo:rerun-if-changed=protos/narwhal.proto");
    println!("qproto");
    tonic_build::compile_protos("../types/narwhal.proto")?;
    Ok(())
}
