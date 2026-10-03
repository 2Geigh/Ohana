package com.nicholasgarcia.ohana.indexer;

import java.nio.file.Path;
import java.sql.Connection;
import java.sql.DriverManager;
import java.sql.PreparedStatement;
import java.sql.ResultSet;
import java.sql.SQLException;
import java.time.LocalDateTime;
import java.time.format.DateTimeFormatter;

import org.jsoup.Jsoup;
import org.jsoup.nodes.Document;
import org.jsoup.nodes.Element;

import org.apache.tika.langdetect.optimaize.OptimaizeLangDetector;
import org.apache.tika.language.detect.LanguageDetector;
import org.apache.tika.language.detect.LanguageResult;

import org.apache.lucene.analysis.Analyzer;
import org.apache.lucene.analysis.standard.StandardAnalyzer;
import org.apache.lucene.document.Field;
import org.apache.lucene.document.StringField;
import org.apache.lucene.document.TextField;
import org.apache.lucene.index.IndexWriter;
import org.apache.lucene.index.IndexWriterConfig;
import org.apache.lucene.index.Term;
import org.apache.lucene.store.Directory;
import org.apache.lucene.store.FSDirectory;

class indexer {

    private static final Analyzer analyzer = new StandardAnalyzer();

    public static final Path INDEX_PATH = Path.of("/data/lucene-index");

    public static void Index(page p) {
        org.apache.lucene.document.Document doc = new org.apache.lucene.document.Document();

        p.GetLanguage();

        try (
                Directory index_dir = FSDirectory.open(INDEX_PATH); IndexWriter writer = new IndexWriter(index_dir, new IndexWriterConfig(analyzer));) {
            doc.add(
                    new StringField(
                            "sql_page_id",
                            Long.toString(p.Sql_id),
                            Field.Store.YES));
            doc.add(
                    new StringField(
                            "sql_site_id",
                            Long.toString(p.Sql_site_id),
                            Field.Store.YES));
            doc.add(
                    new StringField(
                            "language",
                            p.Language,
                            Field.Store.YES));
            doc.add(
                    new TextField(
                            "title",
                            p.Title,
                            Field.Store.NO));
            doc.add(
                    new TextField(
                            "description",
                            p.Description,
                            Field.Store.NO));

            doc.add(
                    new TextField(
                            "body",
                            p.GetTextContent(),
                            Field.Store.NO));

            writer.updateDocument(
                    new Term("sql_page_id", Long.toString(p.Sql_id)),
                    doc);

            writer.commit();
        } catch (Exception e) {
            System.err.println("index failed: " + e.getMessage());
        }

    }
}

class page {

    public Long Sql_id;
    public Long Sql_site_id;
    public String Title;
    public String Description;
    public String Url;
    // public String TextContent;
    public String ResponseBody;
    public String Language;

    private static final LanguageDetector detector = new OptimaizeLangDetector().loadModels();

    private LanguageResult languageResult;

    public void GetLanguage() {
        languageResult = detector.detect(
                GetTextContent()
        );

        if (languageResult.getLanguage().length() < 2) {
            Language = "xx";
            return;
        }

        if (languageResult.isReasonablyCertain()) {
            Language = "xx";
            return;
        }

        if (languageResult.isUnknown()) {
            Language = "xx";
            return;
        }

        Language = languageResult.getLanguage().substring(0, 2);
    }

    public String GetTextContent() {
        Document document = Jsoup.parse(ResponseBody);

        Element root = document.body();
        if (root == null) {
            root = document;
        }

        String text = root.text();
        if (text == null) {
            text = "";
        }

        String normalized_text = text
                .replace('\u00A0', ' ')
                .replaceAll("\\s+", " ")
                .trim();

        return normalized_text;
    }

    public page(
            long sqlId,
            long sqlSiteId,
            String url,
            String title,
            String description,
            String responseBody) {

        this.Sql_id = sqlId;
        this.Sql_site_id = sqlSiteId;
        this.Url = url;
        this.Title = title;
        this.Description = description;
        this.ResponseBody = responseBody;
    }

    public void VerifyDbQueryResults(Connection conn) throws Exception {
        String err_focus = "";
        if (Sql_id == null) {
            err_focus = "page_id";
        } else if (Sql_site_id == null) {
            err_focus = "site_id";
        } else if (Url == null) {
            err_focus = "link";
        } else if (ResponseBody == null) {
            err_focus = "response_body";
        } else if (Title == null) {
            err_focus = "title";
        } else if (Description == null) {
            err_focus = "description";
        }

        boolean areResultsValid = err_focus.isEmpty();

        if (areResultsValid) {
            return;
        }

        try (
                conn; PreparedStatement s = conn.prepareStatement(
                        "DELETE FROM indexer_queue WHERE id = ( SELECT id FROM indexer_queue ORDER BY id DESC LIMIT 1\n);");) {
            s.executeUpdate();
        } catch (Exception e) {
            System.err.println("dequeue page from indexer_queue failed: " + e.getMessage());
        }

        throw new Exception("invalid " + err_focus + " returned from database query");
    }

}

public class App {

    private static void logIteration(page p) {
        System.out.println(
                LocalDateTime.now().format(DateTimeFormatter.ofPattern("yyyy/MM/dd HH:mm:ss")) + " [" + p.Url + "] " + p.Language
        );
    }

    private static void updateDatabase(
            Connection conn,
            page p
    ) throws Exception {
        try {
            conn.setAutoCommit(false); // Begin transaction
        } catch (SQLException e) {
            throw new Exception("DELETE FROM keywords failed: " + e.getMessage());
        }

        try (
                PreparedStatement stmt = conn.prepareStatement(
                        "DELETE FROM keywords WHERE page_id = ?;");) {
            stmt.setObject(1, p.Sql_id);
            stmt.executeUpdate();
        } catch (SQLException e) {
            conn.rollback();
            throw new Exception("DELETE FROM keywords failed: " + e.getMessage());
        }

        try (
                PreparedStatement stmt = conn.prepareStatement(
                        "UPDATE pages SET page_text = ?, date_last_indexed = ?, text_language = ? WHERE id = ?;");) {
            // stmt.setObject(1, "embedding");
            stmt.setObject(1, p.GetTextContent());
            stmt.setObject(2, new java.sql.Timestamp(System.currentTimeMillis()));
            stmt.setObject(3, p.Language);
            stmt.setObject(4, p.Sql_id);
            int rows_updated = stmt.executeUpdate();
            if (rows_updated == 0) {
                throw new Exception("execute UPDATE pages failed");
            }
        } catch (SQLException e) {
            conn.rollback();
            throw new Exception("UPDATE pages failed: " + e.getMessage());
        }

        try (
                PreparedStatement stmt = conn.prepareStatement(
                        "DELETE FROM indexer_queue WHERE page_id = ?;");) {
            stmt.setObject(1, p.Sql_id);
            int rows_updated = stmt.executeUpdate();
            if (rows_updated == 0) {
                throw new Exception("No row found to delete in indexer_queue with page_id = " + p.Sql_id);
            }
        } catch (SQLException e) {
            conn.rollback();
            throw new Exception("DELETE FROM indexer_queue failed: " + e.getMessage());
        } catch (Exception e) {
            conn.rollback();
            System.err.println("DELETE FROM indexer_queue failed: " + e.getMessage());
        }

        try (
                PreparedStatement stmt = conn.prepareStatement(
                        "INSERT INTO ranking_engine_queue (page_id, site_id) VALUES (?, ?);");) {
            stmt.setObject(1, p.Sql_id);
            stmt.setObject(2, p.Sql_site_id);
            int rows_updated = stmt.executeUpdate();
            if (rows_updated == 0) {
                throw new Exception("INSERT INTO ranking_engine_queue failed");
            }
        } catch (SQLException e) {
            conn.rollback();
            throw new Exception("INSERT INTO ranking_engine_queue failed: " + e.getMessage());
        } catch (Exception e) {
            conn.rollback();
            throw new Exception("INSERT INTO ranking_engine_queue failed: " + e.getMessage());
        }

        try {
            conn.commit();
        } catch (SQLException e) {
            conn.rollback();
            throw new Exception("commit transaction failed: " + e.getMessage());
        }

        try {
            conn.setAutoCommit(true);
        } catch (SQLException e) {
            throw new Exception("enable autocommit (disable transaction) failed: " + e.getMessage());
        }
    }

    public static void main(String[] args) {

        final String DB_HOST = System.getenv("DB_HOST");
        final String DB_PASSWORD = System.getenv("DB_PASSWORD");
        final String DB_USERNAME = System.getenv("DB_USERNAME");
        final String DB_CONTAINER_PORT = System.getenv("DB_CONTAINER_PORT");
        final String DB_NAME = System.getenv("DB_NAME");
        final String JDBC_URL = "jdbc:postgresql://" + DB_HOST + ":" + DB_CONTAINER_PORT + "/" + DB_NAME;

        try (Connection connection = DriverManager.getConnection(JDBC_URL, DB_USERNAME, DB_PASSWORD);) {
            while (true) {

                page p;

                try (
                        PreparedStatement stmt
                        = connection
                                .prepareStatement(
                                        "SELECT indexer_queue.page_id, indexer_queue.site_id, pages.link, pages.response_body, pages.description, pages.title FROM indexer_queue LEFT JOIN pages ON pages.id = indexer_queue.page_id ORDER BY indexer_queue.id ASC LIMIT 1;");) {
                            ResultSet result = stmt.executeQuery();

                            boolean isIndexerQueueEmpty = !(result.next());
                            if (isIndexerQueueEmpty) {
                                Thread.sleep(1000);
                                // TODO: then get the page that has been indexed the longest time ago.
                                continue;
                            }

                            p = new page(
                                    result.getLong("page_id"),
                                    result.getLong("site_id"),
                                    result.getString("link"),
                                    result.getString("title"),
                                    result.getString("description"),
                                    result.getString("response_body")
                            );

                        } catch (SQLException e) {
                            throw new Exception("prepare stmt failed: " + e.getMessage());
                        }

                        try {
                            p.VerifyDbQueryResults(connection);
                        } catch (Exception e) {
                            throw new Exception("some aspect of the page data initialization failed: " + e.getMessage());
                        }

                        indexer.Index(p);
                        // TODO: Run [sentence-transformers/all-MiniLM-L6-v2]("https://huggingface.co/sentence-transformers/all-MiniLM-L6-v2") in ONNX runtime
                        updateDatabase(connection, p);

                        logIteration(p);
                        System.gc();
            }
        } catch (Exception e) {
            System.err.println(e.getMessage());
            e.printStackTrace();
            System.exit(1);
        }
    }
}
