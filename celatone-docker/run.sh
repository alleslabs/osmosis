#!/bin/bash

osmosisd init "$1" --chain-id localosmosis

cp /chain/celatone-docker/"$1"/priv_validator_key.json ~/.osmosisd/config/priv_validator_key.json
cp /chain/celatone-docker/"$1"/node_key.json ~/.osmosisd/config/node_key.json

cp /chain/celatone-docker/config/genesis.json ~/.osmosisd/config/genesis.json
cp /chain/celatone-docker/config/app.toml ~/.osmosisd/config/app.toml
cp /chain/celatone-docker/config/config.toml ~/.osmosisd/config/config.toml

sleep 60
osmosisd start --rpc.laddr tcp://0.0.0.0:26657 --with-emitter LocalOsmosis@172.18.0.31:9092
