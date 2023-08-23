import json
import click
import sys

from confluent_kafka import Consumer, TopicPartition
from loguru import logger
from sqlalchemy import create_engine
from urllib.parse import urlparse


from .cli import cli
from .db import tracking
from .handler import Handler


def make_confluent_config(servers, username, password, topic_id):
    return {
        "bootstrap.servers": servers,
        "group.id": topic_id + "-flusher",
        "security.protocol": "SASL_SSL",
        "sasl.mechanisms": "PLAIN",
        "sasl.username": username,
        "sasl.password": password,
        "auto.offset.reset": "smallest",
        "enable.auto.offset.store": False,
    }


@cli.command()
@click.option(
    "--db",
    help="Database URI connection string.",
    default="localhost:5432/postgres",
    show_default=True,
)
@click.option(
    "-s",
    "--servers",
    help="Kafka bootstrap servers.",
    default="localhost:9092",
    show_default=True,
)
@click.option(
    "-u",
    "--username",
    help="Username",
    default="username",
    show_default=True,
)
@click.option(
    "-p",
    "--password",
    help="Password",
    default="password",
    show_default=True,
)
@click.option(
    "-s",
    "--servers",
    help="Kafka bootstrap servers.",
    default="localhost:9092",
    show_default=True,
)
@click.option(
    "-tid",
    "--topic-id",
    help="Topic ID",
    default="topic id",
    show_default=True,
)
@click.option("-e", "--echo-sqlalchemy", "echo_sqlalchemy", is_flag=True)
def sync(db, servers, username, password, echo_sqlalchemy, topic_id):
    """Subscribe to Kafka and push the updates to the database."""
    logger.configure(handlers=[{"sink": sys.stderr, "level": "INFO", "serialize": True}])
    # Set up Kafka connection
    engine = create_engine("postgresql+psycopg2://" + db, echo=echo_sqlalchemy)
    tracking_info = engine.execute(tracking.select()).fetchone()
    consumer = Consumer(make_confluent_config(servers, username, password, topic_id))
    partition = TopicPartition(topic_id, 0, tracking_info.kafka_offset + 1)
    consumer.assign([partition])
    consumer.seek(partition)

    # try log an error and a warning
    logger.error("This is an error")
    logger.warning("This is a warning")

    while True:
        with engine.begin() as conn:
            while True:
                handler = Handler(conn)
                messages = consumer.consume(num_messages=1, timeout=2.0)
                if len(messages) == 0:
                    continue
                msg = messages[0]
                key = msg.key().decode()
                value = json.loads(msg.value())
                if key == "COMMIT":
                    conn.execute(tracking.update().values(kafka_offset=msg.offset()))
                    logger.info(
                        "Committed at block {height} and Kafka offset {offset}",
                        height=value["height"],
                        offset=msg.offset(),
                    )
                    break
                getattr(handler, "handle_" + key.lower())(value)
