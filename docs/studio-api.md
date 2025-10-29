# Stick Studio API

Studio provides CORS-enabled endpoints for a future UI client to interact with the Stick system without coupling to the agent.

## Configuration

- Env vars:
  - `STICK_STUDIO_PORT` (default `8080`)
  - `STICK_STUDIO_ALLOWED_ORIGINS` (default `*`)
  - `STICK_ENV` (default `development`)
- Viper keys (optional): `studio.port`, `studio.allowed_origins`, `studio.env`

## Endpoints

- `GET /api/health`
  - Returns `{status, env, version}`
  - Status codes: `200`

- `GET /api/conversations?limit=50&offset=0`
  - Returns `{data: Conversation[], limit, offset}`
  - Status codes: `200`, `500`

- `GET /api/conversations/:id`
  - Query `with_messages=true` to include messages
  - Returns `{conversation, messages?}` or `Conversation`
  - Status codes: `200`, `400`, `404`

- `GET /api/conversations/:id/messages`
  - Returns `{data: Message[]}`
  - Status codes: `200`, `400`, `500`

- `GET /api/usage/:conversationId`
  - Returns `{data: Usage[]}`
  - Status codes: `200`, `400`, `500`

- `GET /api/functions`
  - Returns `{data: [{name, min_args, max_args}]}`
  - Status codes: `200`

- `POST /api/functions/:name/execute`
  - Body: `{ "args": ["..."] }`
  - Returns `{result: string}`
  - Status codes: `200`, `400`

## Error Format

```
{
  "error": {
    "code": "string",
    "message": "string"
  }
}
```

## Notes

- CORS is configurable via `AllowedOrigins`; methods allowed: `GET, POST, OPTIONS`.
- Standard functions are exposed via registry; agent integration remains separate.
- Logging uses custom request logs; panics are recovered by middleware.