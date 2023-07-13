FROM golang:1.20-bullseye

WORKDIR /chain
COPY . /chain


COPY ./celatone-docker/run.sh .

RUN make install

CMD osmosisd start --rpc.laddr tcp://0.0.0.0:26657
