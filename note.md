cd Docker

docker compose up -d


cd benchmark
attach
pip install -r requirements.txt
fab local

cd bin
./benchmark_client --nodes http://localhost:4001 --rate 20 --size 10 http://localhost:4001