<?php

declare(strict_types=1);

require_once __DIR__ . "/../src/Router.php";

$PATH = parse_url($_SERVER['REQUEST_URI'], PHP_URL_PATH);

$router = new Router;

$router->Add("/", function () {
    echo "welcome to / !";
});

$router->Add("/search", function () {
    switch ($_SERVER['REQUEST_METHOD']) {
        case 'GET':
            $isValidQueryParam = key_exists('q', $_GET);
            if (!$isValidQueryParam) {
                echo "Welcome to the search page, without the clutter of the homepage.";
                return;
            }

            $query = $_GET['q'];

            $isValidQuery = $query . trim($query) != '';
            if (!$isValidQuery) {
                header('Location: /search');
                return;
            }

            echo "You searched for: `{$query}'.\n";
            echo "Here are your search results...";
            return;

        default:
            http_response_code(405);
            return;
    }
});

$router->Dispatch($PATH);
