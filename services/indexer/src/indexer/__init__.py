import asyncio
import asyncpg
from bs4 import BeautifulSoup
from datetime import datetime
import importlib.util
import logging
import os
from polyglot.detect import Detector
from polyglot.detect.base import UnknownLanguage, Language
from sentence_transformers import SentenceTransformer
import spacy
import subprocess
import sys
import time
import traceback


logging.basicConfig(
    level=logging.INFO,
    stream=sys.stdout,
    format="%(levelname)s %(asctime)s %(name)s: %(message)s",
    force=True,
)

logger = logging.getLogger(__name__)

def encodeUtf8(text: str) -> str:
    if not isinstance(text, str):
        text = str(text)

    # Replace characters that cannot be encoded as valid UTF-8.
    text = text.encode("utf-8", errors="replace").decode("utf-8")

    # Remove NUL bytes and Unicode surrogate characters.
    text = text.replace("\x00", "")
    text = "".join(
        character for character in text if not 0xD800 <= ord(character) <= 0xDFFF
    )

    return text


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
                LEFT JOIN pages
                ON pages.id = indexer_queue.page_id
                ORDER BY pages.id
                ASC
                LIMIT 1;"""
            )

            if len(values) < 1:
                await asyncio.sleep(1)
                continue

            pageId = values[0]["page_id"]
            siteId = values[0]["site_id"]

            
            url = values[0]["link"]
            print(url)

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
            stripped_text = text.strip()  # Removes leading and trailing whitespace
            trimmed_text = " ".join(
                stripped_text.split()  # Removes excessive in-text whitespace
            )
            CLEANED_TEXT = encodeUtf8(trimmed_text)
            print(CLEANED_TEXT)

            #########################################
            ########## KEYWORD EXTRACTION ###########
            #########################################

            # Determine page language
            try:
                page_language = Detector(CLEANED_TEXT).language
            except Exception as exc:
                logger.error(
                    f"detect page {pageId} language failed: {exc}",
                )
                page_language = None
            print(page_language)

            SPACY_MODEL_NAMES = {
                "Multilingual": "xx_sent_ud_sm",
                "Catalan": "ca_core_news_trf",
                "Chinese": "zh_core_web_trf",
                "Croatian": "hr_core_news_lg",
                "Danish": "da_core_news_trf",
                "Dutch": "nl_core_news_lg",
                "English": "en_core_web_trf",
                "Finnish": "fi_core_news_lg",
                "French": "fr_dep_news_trf",
                "German": "de_dep_news_trf",
                "Greek": "el_core_news_lg",
                "Italian": "it_core_news_lg",
                "Japanese": "ja_core_news_trf",
                "Korean": "ko_core_news_lg",
                "Lithuanian": "lt_core_news_lg",
                "Macedonian": "mk_core_news_lg",
                "Polish": "pl_core_news_lg",
                "Portuguese": "pt_core_news_lg",
                "Romanian": "ro_core_news_lg",
                "Russian": "ru_core_news_lg",
                "Slovenian": "sl_core_news_trf",
                "Spanish": "es_dep_news_trf",
                "Swedish": "sv_core_news_lg",
                "Ukrainian": "uk_core_news_trf",
            }  # based on https://spacy.io/usage#quickstart

            model_name = SPACY_MODEL_NAMES["Multilingual"]
            if page_language != None and page_language.name in SPACY_MODEL_NAMES:
                model_name = SPACY_MODEL_NAMES[page_language.name]
            print("MODEL NAME", model_name)

            if importlib.util.find_spec(model_name) is None:
                
                subprocess.check_call(
                    [
                        sys.executable,
                        "-m",
                        "spacy",
                        "download",
                        model_name,
                    ]
                )

            # nlp = spacy.load(model_name)

            # # Process the text
            # doc = nlp(text)

            # # Extract nouns and proper nouns as potential keywords
            # keywords = []
            # for token in doc:
            #     if token.pos_ in ["NOUN", "PROPN"] and not token.is_stop:
            #         keywords.append(token.text)

            # # Extract noun chunks (phrases like "data science")
            # noun_chunks = [chunk.text for chunk in doc.noun_chunks]

            # print("Important Nouns:", set(keywords))
            # print("Noun Chunks:", noun_chunks[:5])  # Show first 5

            #########################################
            ########## VECTORIZE PAGE TEXT ##########
            #########################################

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

            page_language_code = "un"
            if page_language != None:
                page_language_code = page_language.code[0:2]

            async with conn.transaction():
                await conn.execute(
                    """UPDATE pages
                    SET
                        embedding = $1,
                        page_text = $2,
                        date_last_indexed = $3,
                        text_language = $4
                    WHERE id = $5;""",
                    embedding_literal,
                    CLEANED_TEXT,
                    datetime.now(),
                    page_language_code,
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
