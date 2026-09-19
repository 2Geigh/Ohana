from datetime import timezone
from datetime import datetime
import asyncio
import asyncpg
import logging
import os
import sys
import time
import traceback
from bs4 import BeautifulSoup
from sentence_transformers import SentenceTransformer


logging.basicConfig(
    level=logging.INFO,
    stream=sys.stdout,
    format="%(levelname)s %(asctime)s %(name)s: %(message)s",
    force=True,
)

logger = logging.getLogger(__name__)


async def indexer() -> None:
    try:
        logger.info("Loading embedding model...")

        start = time.perf_counter()
        model = SentenceTransformer(
            "sentence-transformers/all-MiniLM-L6-v2", device="cpu"
        )

        logger.info("Model loaded in %.1f seconds", time.perf_counter() - start)

        conn = await asyncpg.connect(
            user=os.getenv("DB_USERNAME"),
            password=os.getenv("DB_PASSWORD"),
            database=os.getenv("DB_NAME"),
            host=os.getenv("DB_HOST"),
        )

        while True:
            values: list = await conn.fetch(
                """SELECT *
                FROM indexer_queue
                ORDER BY id
                ASC
                LIMIT 1;"""
            )

            if len(values) < 1:
                await asyncio.sleep(1)
                continue

            pageId = values[0]["page_id"]
            siteId = values[0]["site_id"]

            values: list = await conn.fetch(
                """SELECT (
                    response_body
                )
                FROM pages
                WHERE id = $1
                LIMIT 1;""",
                pageId,
            )

            if len(values) < 1:
                continue

            result: asyncpg.protocol.record.Record = values[0]
            html = next(result.values())

            soup = BeautifulSoup(html, "lxml")
            soup.prettify()
            text = soup.get_text()
            trimmed_text = text.strip()  # Removes leading and trailing whitespace
            CLEANED_TEXT = " ".join(
                trimmed_text.split()  # Removes excessive in-text whitespace
            )

            # TODO: Get keywords from CLEANED_TEXT

            logger.info(
                "Encoding page %s: %d characters",
                pageId,
                len(CLEANED_TEXT),
            )

            start = time.perf_counter()

            embedding = model.encode(
                [CLEANED_TEXT],
                show_progress_bar=True,
                convert_to_numpy=True,
            )

            logger.info(
                "Embedding completed in %.1f seconds; shape=%s",
                time.perf_counter() - start,
                embedding.shape,
            )

            # Convert shape (1, N) to shape (N,)
            # where N is the number of dimensions
            embedding_values = embedding[0].tolist()

            # Convert to pgvector text syntax
            embedding_literal = (
                "[" + ",".join(str(float(value)) for value in embedding_values) + "]"
            )

            async with conn.transaction():
                await conn.execute(
                    """UPDATE pages
                    SET
                        embedding = $1,
                        page_text = $2,
                        date_last_indexed = $3
                    WHERE id = $4;""",
                    embedding_literal,
                    CLEANED_TEXT,
                    datetime.now(),
                    pageId,
                )

                await conn.execute(
                    """DELETE FROM indexer_queue
                    WHERE page_id = $1;""",
                    pageId,
                )

                await conn.execute(
                    """INSERT INTO ranking_engine_queue
                    (page_id, site_id)
                    VALUES ($1, $2);""",
                    pageId,
                    siteId,
                )

                # TODO: Append webpage id to arrays in each keyword's key-value db entry

    except Exception as error:
        print(f"Error: {error}")
        traceback.print_exc()

    finally:
        if conn is not None:
            await conn.close()


def main():
    asyncio.run(indexer())


if __name__ == "__main__":
    main()
