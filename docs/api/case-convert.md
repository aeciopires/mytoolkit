<!-- TOC -->

- [Case Converter — REST API](#case-converter--rest-api)
  - [Request](#request)
  - [Success response (200)](#success-response-200)
  - [Error response (400)](#error-response-400)

<!-- TOC -->

# Case Converter — REST API

`POST /api/v1/tools/case-convert`

## Request

```json
{ "input": "hello world. this IS a test!", "options": { "mode": "sentence" } }
```

`options.mode`: `sentence`, `upper`, `lower`, `title`, `mixed`, `inverse`.

## Success response (200)

```json
{
  "success": true,
  "data": { "output": "Hello world. This is a test!" },
  "meta": { "tool": "case-convert", "duration_ms": 0.03 }
}
```

## Error response (400)

```json
{ "success": false, "error": { "code": "INVALID_OPTION", "message": "mode must be one of [sentence upper lower title mixed inverse], got bogus" } }
```

## Workflow

```mermaid
flowchart LR
    A[Client: CLI / Web / REST] -->|input, mode| B[handlers.Wrap generic handler]
    B --> C[caseconvert.Convert]
    C --> D{mode}
    D -->|sentence| E[sentenceCase]
    D -->|upper/lower| F[strings.ToUpper/ToLower]
    D -->|title| G[titleCase]
    D -->|mixed| H[mixedCase: alternate by position]
    D -->|inverse| I[inverseCase: swap own case]
    E --> J[shared success envelope]
    F --> J
    G --> J
    H --> J
    I --> J
    D -->|invalid mode| K[apperr.Error 400 INVALID_OPTION] --> L[shared error envelope]
```
