# Ohana 🌿

A language-specific search engine for the living web.

## Motivation

This search engine is designed to do two things:

1. Facilitate search beyond mainstream platforms (Reddit, Weibo, _etc_) into the
   re-fledgling [indie web](https://en.wikipedia.org/wiki/IndieWeb) (_e.g._
   personal websites, Neocities pages, niche forums, _etc_).
2. Allow for language-specific filtering of search results.

In other words, this tool helps you browse the living web in your target
language.

## Timeline

### v 1.0

- [ ] Fully functional web crawler
  - [x] Non-pornographic domain blocklist
  - [ ] Upper bounds on site quantity based on arbitrary target
  - [x] Embedded SQLite queue for BFS
- [ ] Implement lexical search via pg_textsearch
- [ ] Implement a minimal client-facing Spring Boot service to query the database and display search results

### v 2.0

- [ ] Basic ranking engine functionality
- [ ] Crawler upgrades
  - [ ] Pornographic domain blocklist
  - [ ] Source domain blacklist from [v2fly/domain-list-community](https://github.com/v2fly/domain-list-community)
  - [ ] Upper bounds on site quantity based on "indie-ness" ranking
- [ ] Client
  - [ ] Divide search results into multiple pages like Google
  - [ ] Add CSS styling
  - [ ] Add "recently added" section to homepage
  - [ ] Add "most visited" section to webpage
  - [ ] Add lightmode-darkmode
  - [ ] Add "featured website" section
  - [ ] Add privacy policy
  - [ ] Add manifesto/about page

### v 3.0

    TBD
