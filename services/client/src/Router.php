<?php

class Router
{
    private $routes = [];

    public function Add(string $path, callable $handler): void
    {
        $this->routes[$path] = $handler;
    }

    public function Dispatch(string $path): void
    {
        $isValidPath = key_exists($path, $this->routes);

        if (!$isValidPath) {
            http_response_code(404);
            return;
        }

        $handler = $this->routes[$path];
        call_user_func($handler);
    }
}
