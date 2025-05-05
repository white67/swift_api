# SWIFT Codes API

This is a RESTful API written in `Go` that provides access to a SWIFT codes database. It supports querying, adding, and deleting banks' SWIFT code data.

## Requirements

- Docker
- Go (optional, only if you want to run or test outside Docker; recommended version 1.24+)

## How to Run the App

1. Clone the repository.
2. Run the following command in the project root:

```bash
docker-compose up --build
```

The API will be available at:

```bash
http://localhost:8080
```

## How to Run Tests

1. Locally (outside Docker)

```bash
go test -v ./internal/model ./internal/parser ./internal/database ./internal/handler ./swift_api/ -short
```

2. Inside Docker

```bash
docker-compose -f docker-compose.test.yml up --build
```

## Implemented endpoints

1. Get details of a SWIFT code

    `GET /v1/swift-codes/{swift-code}`

    Returns details about the SWIFT code. If it is a headquarter, includes its branches.

2. Get all SWIFT codes for a country

    `GET /v1/swift-codes/country/{countryISO2}`
    
    Returns all SWIFT codes for a given country (headquarters and branches).

3. Add a new SWIFT code

    `POST /v1/swift-codes`

    Creates a new SWIFT code entry.

    Request JSON format example:
```json
{
  "address": "Main Street 123, Warsaw",
  "bankName": "Test Bank",
  "countryISO2": "PL",
  "countryName": "Poland",
  "isHeadquarter": true,
  "swiftCode": "TESTPL12XXX"
}
```

4. Delete a SWIFT code

    `DELETE /v1/swift-codes/{swift-code}`

    Deletes the entry matching the given SWIFT code.