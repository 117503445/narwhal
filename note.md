cd Docker

docker compose up -d


cd benchmark
attach
pip install -r requirements.txt
fab local

cd bin
./bin/benchmark_client --nodes http://localhost:4001 --rate 20 --size 10 http://localhost:4001

docker build -t narwhal-builder -f ./Docker/Dockerfile.builder .
docker run -it --rm -v $PWD/Cargo.toml:/workspace/Cargo.toml -v $PWD/Cargo.lock:/workspace/Cargo.lock -v $PWD/consensus:/workspace/consensus -v $PWD/crates:/workspace/crates -v $PWD/sui-execution:/workspace/sui-execution -v $PWD/narwhal:/workspace/narwhal -v $PWD/external-crates:/workspace/external-crates -v $PWD/.target-docker:/workspace/target sui-builder bash


docker run -it --rm -v $PWD:/workspace sui-builder bash
cargo build --target-dir docker-target --features="benchmark" --bin node --bin benchmark_client 

rustup default nightly

[source.crates-io]
replace-with = 'mirror'

[source.mirror]
registry = "https://mirrors.tuna.tsinghua.edu.cn/git/crates.io-index.git"

code --install-extension rust-lang.rust-analyzer@0.3.1394