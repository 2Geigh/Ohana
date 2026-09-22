package com.nicholasgarcia.ohana.indexer;

import java.sql.Connection;
import java.sql.DriverManager;
import java.sql.PreparedStatement;
import java.sql.ResultSet;
import java.sql.SQLException;
import java.sql.Time;
import java.time.Duration;

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

                boolean isResultInvalid = PAGE_ID == -1 || SITE_ID == -1 || PAGE_URL.equals("");
                if (isResultInvalid) {
                    throw new Exception("invalid row values returned from database query");
                }

                System.out.println(RESPONSE_BODY);

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
