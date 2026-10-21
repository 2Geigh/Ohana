package com.nicholasgarcia.ohana.client;

import org.springframework.boot.SpringApplication;
import org.springframework.boot.autoconfigure.SpringBootApplication;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RestController;

@RestController

@SpringBootApplication
public class App {

        @RequestMapping ("/")
        public String home() {
                return "Welcome to / !";
        }

        public static void main(String[] args) {
                SpringApplication.run(App.class, args);
        }
}
