FROM golang:1.19-buster

WORKDIR /chain
COPY . /chain


COPY ./celatone-docker/run.sh .

RUN make install

CMD osmosisd start --rpc.laddr tcp://0.0.0.0:26657
