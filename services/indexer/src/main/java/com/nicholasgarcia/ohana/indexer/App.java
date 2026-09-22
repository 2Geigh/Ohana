package com.nicholasgarcia.ohana.indexer;

import java.sql.Connection;
import java.sql.DriverManager;
import java.sql.SQLException;

public class App {

    public static void main(String[] args) {

        final String DB_HOST = System.getenv("DB_HOST");
        final String DB_PASSWORD = System.getenv("DB_PASSWORD");
        final String DB_USERNAME = System.getenv("DB_USERNAME");
        final String DB_CONTAINER_PORT = System.getenv("DB_CONTAINER_PORT");
        final String DB_NAME = System.getenv("DB_NAME");
        final String JDBC_URL = "jdbc:postgresql://" + DB_HOST + ":" + DB_CONTAINER_PORT + "/" + DB_NAME;

        Connection connection;
        try {
            System.out.println("Connecting to PostgreSQL...");
            connection = DriverManager.getConnection(JDBC_URL, DB_USERNAME, DB_PASSWORD);
            System.out.println("Connected to PostgresSQL successfully.");

            connection.close();
        } catch (SQLException e) {
            System.out.println("connect to postgresql failed: " + e.getMessage());
            e.printStackTrace();
            System.exit(1);
        }
    }
}
