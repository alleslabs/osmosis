#!/bin/bash

DIR=$(dirname "$0")

docker compose up -d --build

sleep 10
