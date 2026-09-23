package com.nicholasgarcia.ohana.indexer;

import java.sql.Connection;
import java.sql.DriverManager;
import java.sql.PreparedStatement;
import java.sql.ResultSet;
import java.sql.SQLException;
import java.time.Duration;
import java.util.ArrayList;
import java.util.Arrays;
import java.util.HashSet;
import java.util.List;
import java.util.Set;
import org.jsoup.Jsoup;
import org.jsoup.nodes.Document;
import org.jsoup.nodes.Element;

public class App {

    public static void main(String[] args) {

        final String DB_HOST = System.getenv("DB_HOST");
        final String DB_PASSWORD = System.getenv("DB_PASSWORD");
        final String DB_USERNAME = System.getenv("DB_USERNAME");
        final String DB_CONTAINER_PORT = System.getenv("DB_CONTAINER_PORT");
        final String DB_NAME = System.getenv("DB_NAME");
        final String JDBC_URL = "jdbc:postgresql://" + DB_HOST + ":" + DB_CONTAINER_PORT + "/" + DB_NAME;

        Connection connection = null;
        try {
            System.out.println("Connecting to PostgreSQL...");
            connection = DriverManager.getConnection(JDBC_URL, DB_USERNAME, DB_PASSWORD);
            System.out.println("Connected to PostgresSQL successfully.");

            while (true) {
                int PAGE_ID = -1, SITE_ID = -1;
                String PAGE_URL = "", RESPONSE_BODY = "";
                PreparedStatement stmt = connection.prepareStatement(
                        "SELECT indexer_queue.page_id, indexer_queue.site_id, pages.link, pages.response_body FROM indexer_queue LEFT JOIN pages ON pages.id = indexer_queue.page_id ORDER BY pages.id ASC LIMIT 1;"
                );
                ResultSet result = stmt.executeQuery();

                boolean isIndexerQueueEmpty = !(result.next());
                if (isIndexerQueueEmpty) {
                    Duration.ofSeconds(1).wait();
                    // TODO: then get the page that has been indexed the longest time ago.

                    stmt.close();
                    result.close();
                    continue;
                }

                PAGE_ID = result.getInt("page_id");
                SITE_ID = result.getInt("site_id");
                PAGE_URL = result.getString("link");
                RESPONSE_BODY = result.getString("response_body");

                boolean isResultInvalid = PAGE_ID == -1 || SITE_ID == -1 || PAGE_URL.equals("") || RESPONSE_BODY.equals("");
                if (isResultInvalid) {
                    throw new Exception("invalid row values returned from database query");
                }

                System.out.println();
                System.out.println("[" + PAGE_URL + "]");

                List<String> chunks = htmlChunkExtractor.extractTextChunks(RESPONSE_BODY);

                System.out.println();
                System.out.println("FINAL RESULTING CHUNKS VVVVVV");
                for (String chunk : chunks) {
                    System.out.println(chunk);
                }

                // TODO: Extract text chunks from the page
                // TODO: Compile the chunks into a single full text source too
                // TODO: Run [sentence-transformers/all-MiniLM-L6-v2]("https://huggingface.co/sentence-transformers/all-MiniLM-L6-v2") in ONNX runtime
                stmt.close();
                result.close();
                break;
            }

        } catch (SQLException e) {
            System.out.println("database/SQL error: " + e.getMessage());
            e.printStackTrace();
            System.exit(e.getErrorCode());
        } catch (Exception e) {
            System.out.println(e.getMessage());
            e.printStackTrace();
            System.exit(1);
        } finally {
            try {
                if (connection != null) {
                    connection.close();
                }
            } catch (SQLException exc) {
                System.out.println("database/SQL error: " + exc.getMessage());
                exc.printStackTrace();
                System.exit(exc.getErrorCode());
            } catch (Exception e) {
                System.out.println(e.getMessage());
                e.printStackTrace();
                System.exit(1);
            }
        }
    }
}

class htmlChunkExtractor {

    private static final Set<String> BLOCK_TAGS = new HashSet<>(Arrays.asList(
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
            "ul",
            "a"
    ));

    private static final Set<String> SKIP_TAGS = new HashSet<>(Arrays.asList(
            "script",
            "style",
            "noscript",
            "template",
            "meta",
            "link"
    ));

    public static List<String> extractTextChunks(String html) {
        if (html == null || html.isBlank()) {
            return List.of();
        }

        Document document = Jsoup.parse(html);

        Element root = document.body();
        if (root == null) {
            root = document;
        }

        List<String> chunks = extractChunks(root, 1000);

        List<String> cleanedChunks = new ArrayList<>();

        for (String chunk : chunks) {
            String cleaned = encodeUtf8(chunk);

            if (!cleaned.isBlank()) {
                cleanedChunks.add(cleaned);
            }
        }

        return cleanedChunks;
    }

    private static List<String> extractChunks(Element element, int maxChars) {
        if (element == null) {
            return List.of();
        }

        if (SKIP_TAGS.contains(element.tagName())) {
            return List.of();
        }

        List<Element> childBlocks = new ArrayList<>();

        for (Element child : element.children()) {
            if (BLOCK_TAGS.contains(child.tagName())) {
                childBlocks.add(child);
            }
        }

        for (Element childBlock : childBlocks) {
            System.out.println();
            System.out.println("CHILD BELOW VVVVVV");
            System.out.println(childBlock);
        }
        boolean isBlockElement = BLOCK_TAGS.contains(element.tagName());
        boolean isChildless = childBlocks.isEmpty();

        if (isBlockElement && isChildless) {
            // Treat the entire element as one text chunk
            String text = normalizeText(element.text());
            if (text.isEmpty()) {
                return List.of();
            }
            return splitIntoChunks(text, maxChars);
        }

        List<String> chunks = new ArrayList<>();

        for (Element child : element.children()) {
            boolean isSkipTag = SKIP_TAGS.contains(child.tagName());
            if (isSkipTag) {
                System.out.println();
                System.out.println("THIS IS A SKIP TAG");
                System.out.println(child);
                continue;
            }

            boolean isBlockTag = BLOCK_TAGS.contains(child.tagName());
            if (!isBlockTag) {
                System.out.println();
                System.out.println("THIS ISN'T A BLOCK TAG");
                System.out.println(child);
                continue;
            }

            System.out.println();
            System.out.println("THIS IS BEING SENT TO TEXT CHUNK EXTRACTION");
            System.out.println(child);
            List<String> text_chunks_in_child = extractChunks(child, maxChars);

            chunks.addAll(text_chunks_in_child);
        }

        //  If no block children produced text,
        //  element's innertext is the fallback
        if (chunks.isEmpty()) {
            String text = normalizeText(element.text());

            if (text.isEmpty()) {
                return List.of();
            }

            chunks.addAll(splitIntoChunks(text, maxChars));
        }

        return chunks;
    }

    /**
     * Normalizes whitespace similarly to:
     *
     * " ".join(tag.get_text(" ", strip=True).split())
     */
    private static String normalizeText(String text) {
        if (text == null) {
            return "";
        }

        return text
                .replace('\u00A0', ' ')
                .replaceAll("\\s+", " ")
                .trim();
    }

    /**
     * Splits long text into chunks without splitting words.
     *
     * The original Python function accepts max_chars but does not currently use
     * it. This implementation applies the limit.
     */
    private static List<String> splitIntoChunks(String text, int maxChars) {
        if (text == null || text.isBlank()) {
            return List.of();
        }

        if (maxChars <= 0 || text.length() <= maxChars) {
            return List.of(text);
        }

        List<String> chunks = new ArrayList<>();
        String[] words = text.split("\\s+");

        StringBuilder currentChunk = new StringBuilder();

        for (String word : words) {
            if (currentChunk.length() == 0) {
                currentChunk.append(word);
                continue;
            }

            int candidateLength = currentChunk.length() + 1 + word.length();

            if (candidateLength <= maxChars) {
                currentChunk.append(' ').append(word);
            } else {
                chunks.add(currentChunk.toString());
                currentChunk.setLength(0);
                currentChunk.append(word);
            }
        }

        if (currentChunk.length() > 0) {
            chunks.add(currentChunk.toString());
        }

        return chunks;
    }

    /**
     * Removes NUL characters and unpaired UTF-16 surrogate characters.
     *
     * Java strings are UTF-16, so this also removes invalid surrogate code
     * units.
     */
    private static String encodeUtf8(String text) {
        if (text == null) {
            return "";
        }

        StringBuilder result = new StringBuilder(text.length());

        for (int i = 0; i < text.length(); i++) {
            char character = text.charAt(i);

            if (character == '\0') {
                continue;
            }

            if (Character.isHighSurrogate(character)) {
                if (i + 1 < text.length()
                        && Character.isLowSurrogate(text.charAt(i + 1))) {

                    result.append(character);
                    result.append(text.charAt(++i));
                }

                continue;
            }

            if (Character.isLowSurrogate(character)) {
                continue;
            }

            result.append(character);
        }

        return result.toString();
    }
}
