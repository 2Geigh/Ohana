import asyncio
import asyncpg
import time
import logging
import sys
import os


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
    try:
        conn = await asyncpg.connect(user=os.getenv("DB_USERNAME"), password=os.getenv("DB_PASSWORD"), database=os.getenv("DB_NAME"), host=os.getenv("DB_HOST"))
        
        while True:
            logger.info("Still indexing web pages...")
            time.sleep(1.25)

    finally:
        conn.close()

if __name__ == '__main__':
    asyncio.run(main())