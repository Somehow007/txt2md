## API Documentation

Base URL: [https://api.example.com/v1](https://api.example.com/v1)

## Authentication

All requests require a Bearer token in the Authorization header.

## Example request:

GET /users HTTP/1.1 Host: api.example.com Authorization: Bearer token123

## Response format:

```
{
  "users": [
    {"id": 1, "name": "Alice"},
    {"id": 2, "name": "Bob"}
  ]
}
```

## Error codes:

1. 400 - Bad Request
2. 401 - Unauthorized
3. 403 - Forbidden
4. 404 - Not Found
5. 500 - Internal Server Error

Rate limiting: 100 requests per minute.