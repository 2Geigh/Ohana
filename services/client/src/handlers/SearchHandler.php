<?php

declare(strict_types=1);

final class Search extends Handler
{
    protected static function get(): void
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
