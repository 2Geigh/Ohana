<?php

declare(strict_types=1);

require_once __DIR__ . "/../src/Router.php";

$PATH = parse_url($_SERVER['REQUEST_URI'], PHP_URL_PATH);

$router = new Router;

$router->Add("/", function () {
    include __DIR__ . "/pages/root/home.php";
});

$router->Add("/search", function () {
    switch ($_SERVER['REQUEST_METHOD']) {
        case 'GET':
            $isValidQueryParam = key_exists('q', $_GET);
            if (!$isValidQueryParam) {
                include __DIR__ . "/pages/search/home.php";
                return;
            }

            $query = $_GET['q'];

            $isValidQuery = $query . trim($query) != '';
            if (!$isValidQuery) {
                header('Location: /search');
                return;
            }

            include __DIR__ . "/pages/search/results.php";
            return;

        default:
            http_response_code(405);
            return;
    }
});

$router->Dispatch($PATH);
