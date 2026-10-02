# Ohana 🌿

A language-specific search engine for the living web.

## Motivation

This search engine is designed to do two things:

1. Facilitate search beyond mainstream platforms (Reddit, Weibo, _etc_) into the re-fledgling [indie web](https://en.wikipedia.org/wiki/IndieWeb) (_e.g._ personal websites, Neocities pages, niche forums, _etc_).
2. Allow for language-specific filtering of search results.

In other words, this tool helps you browse the living web in your target language.

## Architecture

### Crawler

Implemented in [Go](https://go.dev).

### Database

Implemented with [Postgresql](https://postgresql.org).

### Indexer

Implemented in [Java]("https://java.com") [25](https://openjdk.org/projects/jdk/25/) via [Maven](https://maven.apache.org/).

### Ranking Engine

Implemented in [Python](https://python.org).

### Query Engine

Implemented in [Java]("https://java.com") [25](https://openjdk.org/projects/jdk/25/) using [Spring](https://spring.io/) via [Gradle](https://gradle.org/).

### Client

Implemented in [PHP](https://php.net).
