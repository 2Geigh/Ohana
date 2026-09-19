-- +goose Up
ALTER TABLE pages
-- https://huggingface.co/sentence-transformers/all-MiniLM-L6-v2
ALTER COLUMN embedding TYPE VECTOR (384);
