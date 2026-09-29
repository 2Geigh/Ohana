<?php

declare(strict_types=1);

$URI = $_SERVER['REQUEST_URI'];
$PATH = parse_url($URI, PHP_URL_PATH);
$URI_QUERY = parse_url($URI, PHP_URL_QUERY);

switch ($PATH) {
    case '/':
        echo 'WELCOME TO THE HOMEPAGE';
        break;

    case '/search':
        if ($URI_QUERY == null) {
            echo '200 bad request: query missing';
            break;
        }

        $search_query = explode("=", $URI_QUERY, 2);
        if (sizeof($search_query) < 2) {
            echo '200 bad request: malformed query';
            break;
        }

        $search_query_key = $search_query[0];
        $search_query_value = $search_query[1];
        if ($search_query_key != "q") {
            echo '200 bad request: invalid query key';
            break;
        }
        if ($search_query_value == "") {
            echo '200 bad request: empty query value';
            break;
        }

        echo "You searched {$search_query[1]}";
        break;

    default:
        echo '404 page not found';
        break;
}

?>