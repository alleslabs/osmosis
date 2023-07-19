make install

# Clear and create a new fresh DB.
dropdb my_db
createdb my_db

kafka-topics --delete --topic test --bootstrap-server  localhost:9092

kafka-topics --create --bootstrap-server localhost:9092 --replication-factor 1 --partitions 1 --topic test


# Init table from flusher
source ./../flusher/venv/bin/activate
python ./../flusher/main.py init beebchain test replay --db localhost:5432/my_db


osmosisd unsafe-reset-all --home ~/.osmosisd
osmosisd start  --with-emitter test@localhost:9092 --rpc.laddr tcp://0.0.0.0:26657 --pruning=nothing
