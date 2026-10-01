<?php

declare(strict_types=1);

require_once PROJECT_ROOT . "/src/Handler.php";

final class Search extends Handler
{

    private static function search(string $sanitized_query): array
    {
        $url = "http://query-engine:5000?q={$sanitized_query}";

        $ch = curl_init($url);
        curl_setopt($ch, CURLOPT_RETURNTRANSFER, true); // return value instead of sending it to stdout

        $response = curl_exec($ch);

        if (curl_error($ch)) {
            throw new RuntimeException(curl_error($ch));
        }

        return $response;
    }

    protected static function get(): void
    {
        $isValidQueryParam = key_exists('q', $_GET);
        if (!$isValidQueryParam) {
            include PROJECT_ROOT . "/public/pages/search/home.php";
            return;
        }

        $query = filter_input(INPUT_GET, 'q', FILTER_SANITIZE_SPECIAL_CHARS);

        $isValidQuery = trim($query) !== '';
        if (!$isValidQuery) {
            header('Location: /search');
            return;
        }

        $results = [];
        $err = null;

        try {
            $results = static::search($query);
        } catch (Exception $e) {
            $err = $e;
            // echo "An error occured Error: {$e}";
        } finally {
            include PROJECT_ROOT . "/public/pages/search/results.php";
        }

    }

}
