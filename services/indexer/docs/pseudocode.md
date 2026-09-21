# Pseudocode

- Connect to Postgres container

- While True:
    - Grab the page from `pages` that has been indexed the longest time ago, so long as it hasn't been indexed within the last `TIME_PERIOD` (this time period, and the crawlers' constants should be kept in an external file that can be accessed by all services (.env?))

    - If there are no webpages found in `pages`, wait/sleep for `INDEXER_POLITENESS_DURATION`

    - Extract all text chunks from the page, and additionally compile them into a single full text source

    - Run [sentence-transformers/all-MiniLM-L6-v2]("https://huggingface.co/sentence-transformers/all-MiniLM-L6-v2") in ONNX runtime
    
    - Use the above model to vectorize the full page text and store the embedding in the pgvector `pages` column

    - Extract keywords from each text chunk (or the whole page, depending on memory usage in testing) using [Apache Lucene]("https://lucene.apache.org/core/")

    - Save keyword-occurences-page_id triplets to database
