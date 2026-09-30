<?php

declare(strict_types=1);

class Search
{

    public static function Handle(): void
    {
        switch ($_SERVER['REQUEST_METHOD']) {
            case 'GET':
                self::get();
                return;

            default:
                http_response_code(405);
                header("Allow: GET");
                return;
        }
    }

    private static function get(): void
    {
        $isValidQueryParam = key_exists('q', $_GET);
        if (!$isValidQueryParam) {
            include PROJECT_ROOT . "/public/pages/search/home.php";
            return;
        }

        $query = $_GET['q'];

        $isValidQuery = $query . trim($query) != '';
        if (!$isValidQuery) {
            header('Location: /search');
            return;
        }

        include PROJECT_ROOT . "/public/pages/search/results.php";
    }

}
