package com.nicholasgarcia.ohana.query_engine;

public class App {

    public String getGreeting() {
        return "Hi, I'm the query engine! Send your queries my way!";
    }

    public static void main(String[] args) {
        System.out.println(new App().getGreeting());
    }
}
