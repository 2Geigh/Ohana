package com.nicholasgarcia.ohana.query_engine;

import org.springframework.boot.SpringApplication;
import org.springframework.boot.autoconfigure.SpringBootApplication;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RestController;

@RestController
@SpringBootApplication
public class App {

    @RequestMapping("/")
    String[] home() {
        String[] results = {"result 1", "result 2", "result 3"};

        System.out.println("We recieved a GET request! Sending results now...");

        return results;
    }

    public static void main(String args[]) {
        SpringApplication.run(App.class, args);
    }
}
