<?php

declare(strict_types=1);

abstract class Handler
{
    protected static $error = null;

    final public static function Handle(): void
    {
        $method = strtoupper($_SERVER['REQUEST_METHOD'] ?? "");

        $isValidMethod = in_array($method, static::$allowedMethods);
        if (!$isValidMethod) {
            static::methodNotAllowed();
            return;
        }

        $handlers = [
            'GET' => [static::class, 'get'],
            'POST' => [static::class, 'post'],
            'PUT' => [static::class, 'put'],
            'DELETE' => [static::class, 'delete'],
            // 'QUERY' => [static::class, 'query'],
            // 'HEAD' => [static::class, 'head'],
            // 'CONNECT' => [static::class, 'connect'],
            // 'OPTIONS' => [static::class, 'options'],
            // 'TRACE' => [static::class, 'trace'],
            // 'PATCH' => [static::class, 'patch'],
        ];

        $handler = $handlers[$method] ?? null;
        if ($handler === null) {
            static::methodNotAllowed();
            return;
        }

        $handler();
    }

    protected static function get(): void
    {
        static::methodNotAllowed();
    }
    protected static function post(): void
    {
        static::methodNotAllowed();
    }
    protected static function put(): void
    {
        static::methodNotAllowed();
    }
    protected static function delete(): void
    {
        static::methodNotAllowed();
    }
    // protected static function query(): void
    // {
    //     static::methodNotAllowed();
    // }
    // protected static function head(): void
    // {
    //     static::methodNotAllowed();
    // }
    // protected static function connect(): void
    // {
    //     static::methodNotAllowed();
    // }
    // protected static function options(): void
    // {
    //     static::methodNotAllowed();
    // }
    // protected static function trace(): void
    // {
    //     static::methodNotAllowed();
    // }
    // protected static function patch(): void
    // {
    //     static::methodNotAllowed();
    // }

    protected static $allowedMethods = ['GET'];

    final protected static function methodNotAllowed(): void
    {
        http_response_code(405);
        header('Allow: ' . implode(', ', static::$allowedMethods));
    }
}