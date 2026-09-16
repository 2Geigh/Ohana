import asyncio
import time
import logging
import sys


logging.basicConfig(
    level=logging.INFO,
    stream=sys.stdout,
    format="%(levelname)s %(asctime)s %(name)s: %(message)s",
    force=True,
)

logger = logging.getLogger(__name__)


# async def runIndexer(conn: connection) -> RuntimeError:
#     while True:
#         # Get item with the oldest time_since_last_indexed
#         cur = conn.cursor()

        

#         # Index it
        

#         logger.info("Still indexing web pages...")
#         time.sleep(1.25)

async def main() -> None:

    # logger.log("Hi!")

    while True:
        logger.info("Still indexing web pages...")
        time.sleep(1.25)

    # try:
        # conn = db.connect()
        # await runIndexer(conn)

    # finally:
        # db.disconnect(conn)

if __name__ == '__main__':
    asyncio.run(main())