# Flusher

## About

**Flusher** is a simple program implemented in Python to consume Kafka messages from a queue and load them into the 
Postgres database

## Setup instructions for Ubuntu users

For Ubuntu users, to install dependencies including Java, Postgres and Kafka

```shell
sudo apt update
sudo apt upgrade --yes

sudo apt install default-jdk
cd $HOME
curl "https://downloads.apache.org/kafka/3.3.2/kafka_2.13-3.3.2.tgz" -o kafka.tgz
mkdir $HOME/kafka && cd $HOME/kafka
tar -xvzf $HOME/kafka.tgz --strip 1
```

Config the downloaded zookeeper and kafka. Edit the file

```shell
vim $HOME/kafka/config/server.properties
```

With the following values

```toml
log.dirs=/home/ubuntu/kafka/logs
delete.topic.enable = true
message.max.bytes = 52428800
replica.fetch.max.bytes = 52428800
```

Next, let's create a system service for zookeeper. Create the service file

```shell
sudo vim /etc/systemd/system/zookeeper.service
```

With the following code

```shell
[Unit]
Requires=network.target remote-fs.target
After=network.target remote-fs.target

[Service]
Type=simple
User=ubuntu
ExecStart=/home/ubuntu/kafka/bin/zookeeper-server-start.sh /home/ubuntu/kafka/config/zookeeper.properties
ExecStop=/home/ubuntu/kafka/bin/zookeeper-server-stop.sh
Restart=on-abnormal

[Install]
WantedBy=multi-user.target
```

In the similar manner, create a service for kafka

```shell
sudo vim /etc/systemd/system/kafka.service
```

With the following code

```shell
[Unit]
Requires=zookeeper.service
After=zookeeper.service

[Service]
Type=simple
User=ubuntu
ExecStart=/bin/sh -c '/home/ubuntu/kafka/bin/kafka-server-start.sh /home/ubuntu/kafka/config/server.properties > /home/ubuntu/kafka/kafka.log 2>&1'
ExecStop=/home/ubuntu/kafka/bin/kafka-server-stop.sh
Restart=on-abnormal

[Install]
WantedBy=multi-user.target
```

Finally, we can start both services

```shell
sudo systemctl enable zookeeper
sudo systemctl enable kafka
sudo systemctl daemon-reload

sudo systemctl start zookeeper
sudo systemctl start kafka
```

## Setup instructions for macOS users

For macOS users, to install dependencies including Java, Postgres and Kafka

```shell
brew cask install java
brew install postgresql
brew install kafka
```
Start the zookeeper service to provide an in-sync view of Kafka cluster, topics and messages

```shell
brew services start zookeeper
```

Start the postgresql service to provide the database server

```shell
brew services start postgresql
```

Start the kafka service to provide a message queue for the indexer node

```shell
brew services start kafka
```

Now make sure python3 venv is installed

```shell
python3 -m pip install --user virtualenv
```

Then run this command to activate the virtual environment and install python dependencies

```shell
python3 -m venv venv && source venv/bin/activate
pip install -r requirements.txt
```
Note that if you encounter an openssl problem while installing dependencies, simply run

```shell
brew install openssl && export LIBRARY_PATH=$LIBRARY_PATH:/usr/local/opt/openssl/lib/
```

## Running Flusher

Before running the flusher, do not forget to activate the python virtual environment

```shell
cd flusher
source venv/bin/activate
```

And use this command to start flusher

```shell
python3 main.py sync --db <DB_URL>
```
