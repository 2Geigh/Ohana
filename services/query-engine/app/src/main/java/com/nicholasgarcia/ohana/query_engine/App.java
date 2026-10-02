package com.nicholasgarcia.ohana.query_engine;

import java.io.IOException;
import java.nio.file.Paths;
import java.sql.Connection;
import java.sql.DriverManager;
import java.sql.PreparedStatement;
import java.sql.ResultSet;
import java.sql.SQLException;
import java.time.Duration;
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
import org.springframework.web.bind.annotation.RestController;

@RestController

@SpringBootApplication
public class App {

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
    public searchResults home() {
        List<String> results = new ArrayList<>();

        System.out.println("We recieved a GET request! Sending results now...");

        final String DB_HOST = System.getenv("DB_HOST");
        final String DB_PASSWORD = System.getenv("DB_PASSWORD");
        final String DB_USERNAME = System.getenv("DB_USERNAME");
        final String DB_CONTAINER_PORT = System.getenv("DB_CONTAINER_PORT");
        final String DB_NAME = System.getenv("DB_NAME");
        final String JDBC_URL = "jdbc:postgresql://" + DB_HOST + ":" + DB_CONTAINER_PORT + "/" + DB_NAME;

        searchResults response = new searchResults();

        try (
                Connection connection = DriverManager.getConnection(JDBC_URL, DB_USERNAME, DB_PASSWORD); PreparedStatement stmt = connection
                .prepareStatement(
                        "SELECT link, title, description FROM pages WHERE id = ? LIMIT 1;");) {

            List<Document> resultDocs = searchIndex("body", "gentoo linux");
            for (Document doc : resultDocs) {

                int sql_id = Integer.parseInt(doc.get("sql_page_id"));
                int sql_site_id = Integer.parseInt(doc.get("sql_site_id"));
                String language = doc.get("language");

                stmt.setObject(1, sql_id);
                stmt.executeQuery();

                ResultSet result = stmt.executeQuery();

                if (!result.next()) {
                    continue;
                }

                String title = result.getString("title");
                String description = result.getString("description");
                String url = result.getString("link");

                results.add(url);
            }
            System.out.println(resultDocs);
            System.out.println(results);

            response.InputtedQuery = "example query";
            response.ProcessedQuery = "processed query";
            response.Results = results;
            response.SearchDuration = Duration.ofNanos(12345);

            System.out.println(response);
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

// class serializer extends ObjectValueSerializer<searchResults> {
//     @Override
//     public void serializeObject(searchResults search_results, JsonGenerator jgen, SerializationContext context) {
//         jgen.writeStartObject();
//         jgen.writeStringProperty("inputted_query", search_results.InputtedQuery);
//         jgen.writeStringProperty("processed_query", search_results.ProcessedQuery);
//         jgen.writeArray(search_results.Results, 0, search_results.Results.length - 1);
//         jgen.writeNumberProperty("search_duration_ns", search_results.SearchDuration.getNano());
//         jgen.writeEndObject();
//     }
// }
class searchResults {

    public String InputtedQuery = "";
    public String ProcessedQuery = "";
    public List<String> Results = new ArrayList<>();
    public Duration SearchDuration;
}
