import click

from .db import metadata, tracking
from .cli import cli
from loguru import logger

from sqlalchemy import create_engine


@cli.command()
@click.argument("chain_id")
@click.argument("topic")
@click.argument("replay_topic")
@click.option(
    "--db",
    help="Database URI connection string.",
    default="localhost:5432/postgres",
    show_default=True,
)
@logger.catch(reraise=True)
def init(chain_id, topic, replay_topic, db):
    """Initialize the database with empty tables and tracking info."""
    engine = create_engine("postgresql+psycopg2://" + db, echo=True)
    metadata.create_all(engine)
    engine.execute(
        tracking.insert(),
        {"chain_id": chain_id, "topic": topic, "replay_topic": replay_topic, "kafka_offset": -1, "replay_offset": -2},
    )
