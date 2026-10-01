<?php

declare(strict_types=1);

abstract class Handler
{
    final public static function Handle(): void
    {
        switch ($_SERVER['REQUEST_METHOD']) {
            case 'GET':
                static::get();
                return;

            case 'POST':
                static::post();
                return;

            case 'QUERY':
                static::query();
                return;

            case 'HEAD':
                static::head();
                return;
            case 'OPTIONS':
                static::options();
                return;

            case 'TRACE':
                static::trace();
                return;

            case 'PUT':
                static::put();
                return;

            case 'DELETE':
                static::delete();
                return;

            case 'CONNECT':
                static::connect();
                return;

            case 'PATCH':
                static::patch();
                return;

            default:
                static::methodNotAllowed();
                break;
        }
    }

    protected static function get(): void
    {
        self::methodNotAllowed();
    }
    protected static function query(): void
    {
        self::methodNotAllowed();
    }
    protected static function head(): void
    {
        self::methodNotAllowed();
    }
    protected static function post(): void
    {
        self::methodNotAllowed();
    }
    protected static function put(): void
    {
        self::methodNotAllowed();
    }
    protected static function delete(): void
    {
        self::methodNotAllowed();
    }
    protected static function connect(): void
    {
        self::methodNotAllowed();
    }
    protected static function options(): void
    {
        self::methodNotAllowed();
    }
    protected static function trace(): void
    {
        self::methodNotAllowed();
    }
    protected static function patch(): void
    {
        self::methodNotAllowed();
    }

    protected static $allowedMethods = ['GET'];

    final protected static function methodNotAllowed(): void
    {
        http_response_code(405);
        header('Allow: ' . implode(', ', static::allowedMethods));
    }
}