package com.nicholasgarcia.ohana.query_engine;

import java.time.Duration;

import org.springframework.boot.SpringApplication;
import org.springframework.boot.autoconfigure.SpringBootApplication;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RestController;

import tools.jackson.databind.ObjectMapper;

@RestController
@SpringBootApplication
public class App {

    @RequestMapping("/")
    public searchResults home() {
        String[] results = {"result 1", "result 2", "result 3"};

        System.out.println("We recieved a GET request! Sending results now...");

        ObjectMapper objectMapper = new ObjectMapper();
        searchResults search_results = new searchResults();

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
