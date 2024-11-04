<h1 align="center">gRPC-Client-Server-CRUD</h1>

## Table of Contents

- [About](#about)
- [Getting Started](#getting_started)
- [Testing](#testing)

## About <a name = "about"></a>

Приложение, которое использует gRPC для реализации операций CRUD (создание, чтение, обновление и удаление) между клиентом и сервером.

## Getting Started <a name = "getting_started"></a>

Иструкция для установки зависимостей, локального тестирования и запуска

### Установка линтера


```bash
make install-golangci-lint
```

### Установка зависимостей

Для начала необходимо скачать [protoc](https://grpc.io/docs/protoc-installation/).

```bash
protoc --version
```
Локальная установка зависимостей в папку /bin

```bash
make install-deps
```
get в go.mod
```bash
make get-deps
```
Генерация Go-кода для gRPC API на основе proto-декларации в файле users.proto
```bash
make generate-users-api
```

## Testing <a name = "testing"></a>

### Запуск линтера для проверки

```bash
make lint
```
## Запуск сервера локально
```bash
go run cmd/server/main.go
```
### Сборка и запуск сервера
```bash
go build -o bin/server cmd/server/main.go
./bin/server
```
## Запуск клиента локально
```bash
go run cmd/client/main.go
```
### Сборка и запуск клиента
```bash
go build -o bin/client cmd/client/main.go
./bin/client
```
## Для тестирования можно использовать [Postman](https://www.postman.com/)
### Примеры запросов:

- UsersV1/Create
```protobuf
{
    "user": {
        "Age": 41,
        "Email": "alexy@example.com",
        "ID": "1",
        "Info": {
            "City": "Helsinki",
            "Street": "some street"
        },
        "Name": "Alexy Laiho"
    }
}
```
- UsersV1/Get
```protobuf
{
    "ID": "1"
}
```
- UsersV1/GetAll
```protobuf
{
    "limit": "100",
    "offset": "0"
}
```
- UsersV1/Update
```protobuf
{
    "ID": "1",
    "user": {
        "Age": {
            "value": 40
        },
        "Email": {
            "value": "alexyCOBHC@example.com"
        },
        "Info": {
            "City": {},
            "Street": {}
        },
        "Name": {
            "value": "Alexy wild-child Laiho"
        }
    }
}
```
- UsersV1/Delete
```protobuf
{
    "ID": "1"
}
```