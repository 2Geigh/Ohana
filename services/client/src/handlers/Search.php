<?php

declare(strict_types=1);

require_once PROJECT_ROOT . "/src/Handler.php";

final class Search extends Handler
{

    private static function search(string $sanitized_query): string
    {
        $url = "http://query-engine:5000?" . http_build_query(
            ['q' => $sanitized_query],
            '',
            '&',
            PHP_QUERY_RFC3986
        );

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

        $results = new searchResults($query);
        $err = null;

        try {
            $response = static::search($query);

            $decoded = json_decode($response, true);
            if (json_last_error() !== JSON_ERROR_NONE) {
                throw new RuntimeException(
                    'Invalid JSON: ' . json_last_error_msg()
                );
            }

            if (key_exists('Results', $decoded)) {
                foreach ($decoded['Results'] as $result) {
                    $toAdd = new searchResult(
                        $result['Title'],
                        $result["Description"],
                        $result["Url"],
                        $result["Language"]
                    );
                    array_push($results->Results, $toAdd);
                }
            } else {
                $results->Results = [];
            }

            if (key_exists('SearchDuration_ns', $decoded)) {
                $results->SearchDuration_ns = (int) $decoded['SearchDuration_ns'];
            } else {
                $results->SearchDuration_ns = null;
            }

            if (key_exists('ProcessedQuery', $decoded)) {
                $results->ProcessedQuery = $decoded['ProcessedQuery'];
            } else {
                $results->ProcessedQuery = $results->InputtedQuery;
            }
        } catch (Exception $e) {
            $err = $e;
        } finally {
            include PROJECT_ROOT . "/public/pages/search/results.php";
        }

    }

}

final class searchResults
{
    public function __construct(
        public string $InputtedQuery,
        public string $ProcessedQuery = '',
        public array $Results = [],
        public int|null $SearchDuration_ns = 0
    ) {
        $this->InputtedQuery = $InputtedQuery;
        $this->ProcessedQuery = $ProcessedQuery;
        $this->Results = $Results;
        $this->SearchDuration_ns = $SearchDuration_ns;
    }

    public function GetDuration_s(): float
    {
        return round($this->SearchDuration_ns / (10 ** 9), 2);
    }
}

final class searchResult
{
    public function __construct(
        public string $Title,
        public string $Description,
        public string $Url,
        public string $Language
    ) {
        $this->Title = $Title;
        $this->Description = $Description;
        $this->Url = $Url;
        $this->Language = $Language;
    }
}