-- +goose Up
ALTER TABLE sites
ADD COLUMN indie_ranking INTEGER;

CREATE EXTENSION IF NOT EXISTS vector;

ALTER TABLE pages
-- Dimensions based on Tencent EVIE model
ADD COLUMN embedding VECTOR (2048);
