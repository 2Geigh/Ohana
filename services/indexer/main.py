import asyncio
import asyncpg
import logging
import os
import sys
import traceback
from bs4 import BeautifulSoup


logging.basicConfig(
    level=logging.INFO,
    stream=sys.stdout,
    format="%(levelname)s %(asctime)s %(name)s: %(message)s",
    force=True,
)

logger = logging.getLogger(__name__)


async def main() -> None:
    try:
        conn = await asyncpg.connect(
            user=os.getenv("DB_USERNAME"),
            password=os.getenv("DB_PASSWORD"), 
            database=os.getenv("DB_NAME"), 
            host=os.getenv("DB_HOST"))
        
        while True:
            
            values: list = await conn.fetch(
                'SELECT * FROM indexer_queue ORDER BY id ASC LIMIT 1;'
            )

            if len(values) < 1:
                await asyncio.sleep(1)
                continue

            pageId = values[0]['page_id']
            siteId = values[0]['site_id']

            values: list = await conn.fetch(
                """
                SELECT (
                    response_body
                )
                FROM pages
                WHERE id = $1
                LIMIT 1;
                """,
                pageId
            )

            if len(values) < 1:
                continue

            result: asyncpg.protocol.record.Record = values[0]
            html = next(result.values())
            
            soup = BeautifulSoup(html, 'lxml')

            PRETTY_HTML = soup.prettify()
            text = soup.get_text()
            trimmed_text = text.strip() # Removes leading and trailing whitespace
            CLEANED_TEXT = ' '.join(trimmed_text.split()) # Removes excessive in-text whitespace

            # Get keywords from CLEANED_TEXT

            # In a single transaction
                # Append webpage id to arrays in each keyword's key-value db entry
                # Delete this page from indexer queue
                # Add this page to ranker queue
    
    except Exception as error:
        print(f"Error: {error}")
        traceback.print_exc()

    finally:
        if conn is not None:
            await conn.close()

if __name__ == '__main__':
    asyncio.run(main())