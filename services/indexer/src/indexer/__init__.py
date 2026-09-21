import asyncio
import asyncpg
from bs4 import BeautifulSoup, Tag
from datetime import datetime
import gc
import importlib.util
import logging
import os
from polyglot.detect import Detector
from sentence_transformers import SentenceTransformer
import spacy
import subprocess
import sys
import time
import traceback
import torch

# Constrain PyTorch thread utilization to prevent memory ballooning in CPU mode.
torch.set_num_threads(1)

logging.basicConfig(
    level=logging.INFO,
    stream=sys.stdout,
    format="%(levelname)s %(asctime)s %(name)s: %(message)s",
    force=True,
)

logger = logging.getLogger(__name__)

BLOCK_TAGS = {
    "html",
    "body",
    "main",
    "article",
    "section",
    "div",
    "header",
    "footer",
    "aside",
    "nav",
    "p",
    "li",
    "blockquote",
    "h1",
    "h2",
    "h3",
    "h4",
    "h5",
    "h6",
    "pre",
    "td",
    "th",
    "dt",
    "dd",
}
SKIP_TAGS = {"script", "style", "noscript", "template", "meta", "link"}
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
}


def encodeUtf8(text: str) -> str:
    if not isinstance(text, str):
        text = str(text)
    text = text.encode("utf-8", errors="replace").decode("utf-8")
    text = text.replace("\x00", "")
    return "".join(
        character for character in text if not 0xD800 <= ord(character) <= 0xDFFF
    )


def extract_chunks(tag, max_chars=1000):
    if not isinstance(tag, Tag) or tag.name in SKIP_TAGS:
        return []

    direct_text = " ".join(
        str(node).strip()
        for node in tag.children
        if not isinstance(node, Tag) and str(node).strip()
    )

    child_blocks = [child for child in tag.find_all(BLOCK_TAGS, recursive=False)]

    if tag.name in BLOCK_TAGS and not child_blocks:
        text = " ".join(tag.get_text(" ", strip=True).split())
        return [text] if text else []

    chunks = []
    for child in tag.children:
        if not isinstance(child, Tag) or child.name in SKIP_TAGS:
            continue
        if child.name in BLOCK_TAGS:
            chunks.extend(extract_chunks(child, max_chars))

    if not chunks:
        text = " ".join(tag.get_text(" ", strip=True).split())
        return [text] if text else []

    return chunks


async def indexer() -> None:
    logger.info("Loading embedding model...")
    start = time.perf_counter()
    model = SentenceTransformer("sentence-transformers/all-MiniLM-L6-v2", device="cpu")
    logger.info("Model loaded in %.1f seconds", time.perf_counter() - start)

    loaded_nlp_models = {}
    conn = None

    try:
        conn = await asyncpg.connect(
            user=os.getenv("DB_USERNAME"),
            password=os.getenv("DB_PASSWORD"),
            database=os.getenv("DB_NAME"),
            host=os.getenv("DB_HOST"),
        )

        while True:
            try:
                print()
                page_start_time = time.time()

                values: list = await conn.fetch(
                    """SELECT indexer_queue.page_id, indexer_queue.site_id, pages.link 
                    FROM indexer_queue
                    LEFT JOIN pages ON pages.id = indexer_queue.page_id
                    ORDER BY pages.id ASC LIMIT 1;"""
                )

                if not values:
                    await asyncio.sleep(1)
                    continue

                pageId = values[0]["page_id"]
                siteId = values[0]["site_id"]
                url = values[0]["link"]
                print(url)

                page_data: list = await conn.fetch(
                    "SELECT response_body FROM pages WHERE id = $1 LIMIT 1;", pageId
                )
                if not page_data:
                    await conn.execute(
                        "DELETE FROM indexer_queue WHERE page_id = $1;", pageId
                    )
                    continue

                html = page_data[0]["response_body"]

                soup = BeautifulSoup(html, "lxml")
                root = soup.body or soup
                text_chunks = extract_chunks(root)
                soup.decompose()

                cleaned_texts = [
                    encodeUtf8(chunk) for chunk in text_chunks if chunk.strip()
                ]
                full_cleaned_text = " ".join(cleaned_texts)
                if not full_cleaned_text.strip():
                    await conn.execute(
                        "DELETE FROM indexer_queue WHERE page_id = $1;", pageId
                    )
                    continue

                try:
                    page_language = Detector(full_cleaned_text[:5000]).language
                except Exception as exc:
                    logger.error(f"detect page {pageId} language failed: {exc}")
                    page_language = None

                model_name = SPACY_MODEL_NAMES["Multilingual"]
                if page_language and page_language.name in SPACY_MODEL_NAMES:
                    model_name = SPACY_MODEL_NAMES[page_language.name]

                if model_name not in loaded_nlp_models:
                    if importlib.util.find_spec(model_name) is None:
                        subprocess.check_call(
                            [sys.executable, "-m", "spacy", "download", model_name]
                        )
                    loaded_nlp_models[model_name] = spacy.load(model_name)
                nlp = loaded_nlp_models[model_name]

                KEY_TERMS_FREQUENCY: dict[str, int] = {}

                for text_chunk in cleaned_texts:
                    chunk_start_time = time.time()

                    doc = nlp(text_chunk)
                    keywords = [
                        token.text
                        for token in doc
                        if token.pos_ in ["NOUN", "PROPN"] and not token.is_stop
                    ]

                    try:
                        noun_chunks = [chunk.text for chunk in doc.noun_chunks]
                    except Exception:
                        noun_chunks = []

                    for term in keywords + noun_chunks:
                        term_lower = term.lower()
                        if term_lower in KEY_TERMS_FREQUENCY:
                            KEY_TERMS_FREQUENCY[term_lower] += 1 
                        else:
                            KEY_TERMS_FREQUENCY[term_lower] = 1 

                    logger.info(f"Indexed in {round(time.time() - chunk_start_time, 2)}s: {keywords + noun_chunks}")

                embed_start = time.time()
                logger.info("Embedding page text...")
                embedding = model.encode(
                    [full_cleaned_text], show_progress_bar=False, convert_to_numpy=True
                )
                embedding_literal = (  # to match pgvector's vector syntax
                    "[" + ",".join(str(float(v)) for v in embedding[0].tolist()) + "]"
                )
                logger.info(f"Embedding completed in {round(time.time() - embed_start, 2)}")

                page_language_code: str = "un"
                if page_language:
                    page_language_code = page_language.code[0:2]

                logger.info(f"Page processed in {round(time.time() - page_start_time, 2)}s")

                async with conn.transaction():
                    tx_start = time.time()
                    print("Executing database transaction...")

                    await conn.execute(
                        "DELETE FROM keywords WHERE page_id = $1;", pageId
                    )

                    await conn.execute(
                        """UPDATE pages
                        SET embedding = $1, page_text = $2, date_last_indexed = $3, text_language = $4
                        WHERE id = $5;""",
                        embedding_literal,
                        full_cleaned_text,
                        datetime.now(),
                        page_language_code,
                        pageId,
                    )

                    await conn.execute(
                        "DELETE FROM indexer_queue WHERE page_id = $1;", pageId
                    )
                    
                    await conn.execute(
                        "INSERT INTO ranking_engine_queue (page_id, site_id) VALUES ($1, $2);",
                        pageId,
                        siteId,
                    )

                    if KEY_TERMS_FREQUENCY:
                        keyword_records = [
                            (k, pageId, v) for k, v in KEY_TERMS_FREQUENCY.items()
                        ]
                        await conn.copy_records_to_table(
                            "keywords",
                            records=keyword_records,
                            columns=["keyword", "page_id", "word_occurences"],
                        )
                    
                    print(f"Transaction completed in {round(time.time() - tx_start, 2)}s")

            except Exception as exc:
                logger.error(f"Error processing page {pageId}: {exc}")
                traceback.print_exc()

            finally:
                gc.collect()

    finally:
        if conn is not None:
            await conn.close()


def main():
    asyncio.run(indexer())


if __name__ == "__main__":
    main()
