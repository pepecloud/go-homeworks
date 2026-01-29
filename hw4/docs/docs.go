// Package docs сгенерирован для Swagger UI. Перегенерировать: swag init -g main.go -o docs
package docs

import "github.com/swaggo/swag"

const docTemplate = `{
    "swagger": "2.0",
    "info": {
        "title": "Items API",
        "version": "1.0",
        "description": "CRUD для заказов и транзакций. Мутирующие запросы требуют JWT (получить через POST /api/login)."
    },
    "host": "localhost:8080",
    "basePath": "/",
    "paths": {
        "/api/login": {
            "post": {
                "tags": ["auth"],
                "summary": "Вход в API",
                "description": "Проверяет логин и пароль, при успехе возвращает JWT для заголовка Authorization.",
                "consumes": ["application/json"],
                "produces": ["application/json"],
                "parameters": [{
                    "in": "body",
                    "name": "body",
                    "required": true,
                    "schema": {
                        "type": "object",
                        "properties": {
                            "login": {"type": "string"},
                            "password": {"type": "string"}
                        }
                    }
                }],
                "responses": {
                    "200": {"description": "OK", "schema": {"type": "object", "properties": {"token": {"type": "string"}}}},
                    "400": {"description": "Неверный запрос"},
                    "401": {"description": "Неверные логин или пароль"}
                }
            }
        },
        "/api/item": {
            "post": {
                "tags": ["items"],
                "summary": "Создать сущность",
                "description": "Создаёт Order или Transaction по JSON. Если есть поле \"date\" — транзакция, иначе заказ. Требуется JWT.",
                "security": [{"BearerAuth": []}],
                "consumes": ["application/json"],
                "produces": ["application/json"],
                "parameters": [{
                    "in": "body",
                    "name": "body",
                    "required": true,
                    "schema": {"type": "object"}
                }],
                "responses": {
                    "201": {"description": "Created"},
                    "400": {"description": "Bad Request"},
                    "401": {"description": "Требуется авторизация"},
                    "409": {"description": "Сущность с таким ID уже есть"}
                }
            }
        },
        "/api/item/{id}": {
            "get": {
                "tags": ["items"],
                "summary": "Получить сущность по ID",
                "description": "Ищет Order или Transaction по id в пути.",
                "produces": ["application/json"],
                "parameters": [{"in": "path", "name": "id", "type": "integer", "required": true}],
                "responses": {
                    "200": {"description": "OK"},
                    "400": {"description": "Bad Request"},
                    "404": {"description": "Not Found"}
                }
            },
            "put": {
                "tags": ["items"],
                "summary": "Обновить сущность",
                "description": "Обновляет Order или Transaction по id. Тело — как при создании. Требуется JWT.",
                "security": [{"BearerAuth": []}],
                "consumes": ["application/json"],
                "produces": ["application/json"],
                "parameters": [
                    {"in": "path", "name": "id", "type": "integer", "required": true},
                    {"in": "body", "name": "body", "required": true, "schema": {"type": "object"}}
                ],
                "responses": {
                    "200": {"description": "OK"},
                    "400": {"description": "Bad Request"},
                    "401": {"description": "Требуется авторизация"},
                    "404": {"description": "Not Found"}
                }
            },
            "delete": {
                "tags": ["items"],
                "summary": "Удалить сущность",
                "description": "Удаляет Order или Transaction по id в пути. Требуется авторизация.",
                "security": [{"BearerAuth": []}],
                "parameters": [{"in": "path", "name": "id", "type": "integer", "required": true}],
                "responses": {
                    "204": {"description": "No Content"},
                    "400": {"description": "Bad Request"},
                    "401": {"description": "Требуется авторизация"},
                    "404": {"description": "Not Found"}
                }
            }
        },
        "/api/items": {
            "get": {
                "tags": ["items"],
                "summary": "Список сущностей",
                "description": "Возвращает все Order и Transaction в одном массиве.",
                "produces": ["application/json"],
                "responses": {"200": {"description": "OK", "schema": {"type": "array"}}}
            }
        }
    },
    "securityDefinitions": {
        "BearerAuth": {
            "type": "apiKey",
            "in": "header",
            "name": "Authorization"
        }
    }
}`

var SwaggerInfo = &swag.Spec{
	Version:          "1.0",
	Host:             "localhost:8080",
	BasePath:         "/",
	Schemes:          []string{},
	Title:            "Items API",
	Description:      "CRUD для заказов и транзакций. Мутирующие запросы требуют JWT (получить через POST /api/login).",
	InfoInstanceName: "swagger",
	SwaggerTemplate:  docTemplate,
}

func init() {
	swag.Register(SwaggerInfo.InstanceName(), SwaggerInfo)
}
