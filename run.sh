trap "rm server;kill 0" EXIT

go build -o server
./server -port=58500 &
./server -port=58501 &
./server -port=58502 &

sleep 2
echo ">>> start test"
curl "http://localhost:9999/api?key=Tom" &
curl "http://localhost:9999/api?key=Tom" &
curl "http://localhost:9999/api?key=Tom" &

wait