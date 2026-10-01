<?php

declare(strict_types=1);

define("PROJECT_ROOT", dirname(__DIR__));

require_once PROJECT_ROOT . "/src/Router.php";
require_once PROJECT_ROOT . "/src/handlers/Search.php";

$PATH = parse_url($_SERVER['REQUEST_URI'], PHP_URL_PATH);

$router = new Router;

$router->Add("/", function () {
    include PROJECT_ROOT . "/public/pages/root/home.php";
});

$router->Add("/search", fn() => Search::Handle());

$router->Dispatch($PATH);
