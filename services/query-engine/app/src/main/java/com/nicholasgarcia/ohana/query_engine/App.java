package com.nicholasgarcia.ohana.query_engine;

import java.io.IOException;
import java.nio.file.Paths;
import java.sql.Connection;
import java.sql.DriverManager;
import java.sql.PreparedStatement;
import java.sql.ResultSet;
import java.sql.SQLException;
import java.time.Duration;
import java.time.LocalDateTime;
import java.time.format.DateTimeFormatter;
import java.util.ArrayList;
import java.util.List;

import org.apache.lucene.analysis.standard.StandardAnalyzer;
import org.apache.lucene.document.Document;
import org.apache.lucene.index.DirectoryReader;
import org.apache.lucene.index.IndexReader;
import org.apache.lucene.index.StoredFields;
import org.apache.lucene.queryparser.classic.ParseException;
import org.apache.lucene.queryparser.classic.QueryParser;
import org.apache.lucene.search.IndexSearcher;
import org.apache.lucene.search.Query;
import org.apache.lucene.search.ScoreDoc;
import org.apache.lucene.search.TopDocs;
import org.apache.lucene.store.Directory;
import org.apache.lucene.store.FSDirectory;
import org.springframework.boot.SpringApplication;
import org.springframework.boot.autoconfigure.SpringBootApplication;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RequestParam;
import org.springframework.web.bind.annotation.RestController;

@RestController

@SpringBootApplication
public class App {

    final private static void logIteration(String query, Duration timespan) {
        System.out.println(
                LocalDateTime.now().format(DateTimeFormatter.ofPattern("yyyy/MM/dd HH:mm:ss"))
                + " Searched \"" + query + "\" in " + timespan.getSeconds() + "s"
        );
    }

    List<Document> searchIndex(String inField, String queryString) {
        StandardAnalyzer analyzer = new StandardAnalyzer();

        try {
            Query query = new QueryParser(inField, analyzer).parse(queryString);

            String indexPath = "/data/lucene-index";
            Directory indexDirectory = FSDirectory.open(Paths.get(indexPath));

            IndexReader indexReader = DirectoryReader.open(indexDirectory);
            IndexSearcher indexSearcher = new IndexSearcher(indexReader);

            TopDocs topDocs = indexSearcher.search(query, 10);
            StoredFields storedFields = indexSearcher.storedFields();

            List<Document> documents = new ArrayList<>();
            for (ScoreDoc scoreDoc : topDocs.scoreDocs) {
                int docId = scoreDoc.doc;
                Document doc = storedFields.document(docId);
                documents.add(doc);
            }
            return documents;

        } catch (IOException e) {
        } catch (ParseException e) {
        } catch (Exception e) {
        }

        return null;
    }

    @RequestMapping("/")
    public searchResults home(@RequestParam String q) {
        List<searchResult> results = new ArrayList<>();

        final String DB_HOST = System.getenv("DB_HOST");
        final String DB_PASSWORD = System.getenv("DB_PASSWORD");
        final String DB_USERNAME = System.getenv("DB_USERNAME");
        final String DB_CONTAINER_PORT = System.getenv("DB_CONTAINER_PORT");
        final String DB_NAME = System.getenv("DB_NAME");
        final String JDBC_URL = "jdbc:postgresql://" + DB_HOST + ":" + DB_CONTAINER_PORT + "/" + DB_NAME;

        String inputQuery = q;
        searchResults response = new searchResults(inputQuery);

        try (
                Connection connection = DriverManager.getConnection(JDBC_URL, DB_USERNAME, DB_PASSWORD); PreparedStatement stmt = connection
                .prepareStatement(
                        "SELECT link, title, description FROM pages WHERE id = ? LIMIT 1;");) {

            List<Document> resultDocs = searchIndex("body", response.InputtedQuery);
            for (Document doc : resultDocs) {

                int sql_id = Integer.parseInt(doc.get("sql_page_id"));
                int sql_site_id = Integer.parseInt(doc.get("sql_site_id"));

                stmt.setObject(1, sql_id);
                stmt.executeQuery();

                ResultSet result = stmt.executeQuery();

                if (!result.next()) {
                    continue;
                }

                results.add(
                        new searchResult(
                                result.getString("title"),
                                result.getString("description"),
                                result.getString("link"),
                                doc.get("language")
                        ));
            }
            response.ProcessedQuery = response.InputtedQuery;
            response.Results = results;
            response.SearchDuration = Duration.ofNanos(12345);

            logIteration(response.ProcessedQuery, response.SearchDuration);
            return response;

        } catch (SQLException e) {
            System.err.println("SQL exception: " + e.getMessage());
            e.printStackTrace();
            return response;
        }
    }

    public static void main(String args[]) {
        SpringApplication.run(App.class, args);
    }

}

class searchResults {

    public String InputtedQuery;
    public String ProcessedQuery = "";
    public List<searchResult> Results = new ArrayList<>();
    public Duration SearchDuration;

    public searchResults(String input_query) {
        this.InputtedQuery = input_query;
        this.ProcessedQuery = input_query;
    }
}

class searchResult {

    public String Title;
    public String Description;
    public String Url;
    public String Language;

    public searchResult(String title, String desc, String url, String language) {
        this.Title = title;
        this.Description = desc;
        this.Url = url;
        this.Language = language;
    }
}
