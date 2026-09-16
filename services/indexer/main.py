import db
import time
import logging
import sys

local_queue: list[str] = []

logging.basicConfig(
    level=logging.INFO,
    stream=sys.stdout,
    format="%(levelname)s %(asctime)s %(name)s: %(message)s",
    force=True,
)

logger = logging.getLogger(__name__)

def index() -> RuntimeError:
    while True:
        logger.info("Still indexing web pages...")
        time.sleep(1.25)

def main() -> None:
    try:
        conn = db.connect()
        index()

    finally:
        db.disconnect(conn)

if __name__ == '__main__':
    main()