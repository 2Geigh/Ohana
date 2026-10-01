package com.nicholasgarcia.ohana.query_engine;

import java.io.IOException;
import java.nio.file.Paths;
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
        }

        return null;
    }

    @RequestMapping("/")
    public searchResults home() {
        String[] results = {"result 1", "result 2", "result 3"};

        System.out.println("We recieved a GET request! Sending results now...");

        List<Document> resultDocs = searchIndex("body", "gentoo linux");
        for (Document doc : resultDocs) {
            String sql_id = doc.get("sql_page_id");
            String sql_site_id = doc.get("sql_site_id");
            String language = doc.get("language");

            

            String title;
            String description;
        }
        System.out.println(resultDocs);

        searchResults response = new searchResults();
        response.InputtedQuery = "example query";
        response.ProcessedQuery = "processed query";
        response.Results = results;
        response.SearchDuration = Duration.ofNanos(12345);

        return response;
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

    public String InputtedQuery;
    public String ProcessedQuery;
    public String[] Results;
    public Duration SearchDuration;
}
