# API Documentation Guide

This guide provides standards and templates for API documentation in the REChain Quantum-CrossAI IDE Engine.

## Documentation Standards

### REST API Endpoints

Each endpoint should include:

1. **Endpoint URL**
2. **HTTP Method**
3. **Description**
4. **Authentication**
5. **Request Parameters**
6. **Request Body** (if applicable)
7. **Response Format**
8. **Error Codes**
9. **Examples**

### Template

```markdown
## Endpoint Name

**URL:** `/api/v1/resource`

**Method:** `GET`

**Description:** Brief description of what this endpoint does.

**Authentication:** Required (Bearer Token)

### Request Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| id | string | Yes | Resource identifier |
| limit | integer | No | Number of items to return (default: 20) |

### Request Body

```json
{
  "name": "string",
  "description": "string",
  "enabled": true
}
```

### Response

**Success (200):**

```json
{
  "status": "success",
  "data": {
    "id": "123",
    "name": "Example",
    "created_at": "2024-01-15T10:30:00Z"
  }
}
```

**Error (404):**

```json
{
  "status": "error",
  "message": "Resource not found",
  "code": "NOT_FOUND"
}
```

### Error Codes

| Code | HTTP Status | Description |
|------|-------------|-------------|
| NOT_FOUND | 404 | Resource does not exist |
| UNAUTHORIZED | 401 | Invalid or missing token |
| VALIDATION_ERROR | 400 | Invalid request data |
```

## GraphQL Documentation

### Schema Documentation

```markdown
## Type: User

Represents a user in the system.

### Fields

| Field | Type | Description |
|-------|------|-------------|
| id | ID! | Unique identifier |
| name | String | User's display name |
| email | String! | User's email address |
| projects | [Project!]! | Projects owned by user |

### Example Query

```graphql
query GetUser($id: ID!) {
  user(id: $id) {
    id
    name
    email
    projects {
      name
      status
    }
  }
}
```
```

## Code Documentation (Go)

### Package Documentation

```go
// Package auth provides authentication and authorization
// functionality for the REChain IDE.
//
// Basic usage:
//
//	authService := auth.NewService(config)
//	token, err := authService.Login(ctx, credentials)
//
// The package supports multiple authentication providers
// including OAuth2 and API keys.
package auth
```

### Function Documentation

```go
// Authenticate validates user credentials and returns an access token.
//
// The token is valid for 24 hours and should be refreshed using
// the RefreshToken method before expiration.
//
// Parameters:
//   - ctx: Context for cancellation and timeouts
//   - username: User's email or username
//   - password: User's password (will be hashed before comparison)
//
// Returns:
//   - *Token: Access token on success
//   - error: ErrInvalidCredentials, ErrAccountLocked, or ErrInternal
//
// Example:
//
//	token, err := auth.Authenticate(ctx, "user@example.com", "password123")
//	if err != nil {
//	    log.Printf("Authentication failed: %v", err)
//	    return
//	}
//	defer token.Release()
func Authenticate(ctx context.Context, username, password string) (*Token, error) {
    // Implementation
}
```

### Type Documentation

```go
// ServiceConfig configures the authentication service.
//
// All fields have sensible defaults, so you only need to set
// the values you want to change.
type ServiceConfig struct {
    // TokenDuration is how long access tokens remain valid.
    // Default: 24 hours
    TokenDuration time.Duration
    
    // MaxLoginAttempts is the number of failed attempts before
    // account is temporarily locked.
    // Default: 5
    MaxLoginAttempts int
    
    // Providers is a list of enabled OAuth2 providers.
    Providers []OAuthProvider
}
```

## SDK Documentation

### TypeScript/JavaScript SDK

```markdown
## Class: IDEClient

Main client for interacting with the REChain IDE API.

### Constructor

```typescript
const client = new IDEClient({
  apiKey: 'your-api-key',
  baseURL: 'https://api.rechain.ai',
  timeout: 30000
});
```

### Methods

#### compile()

Compile smart contract source code.

```typescript
async compile(params: CompileParams): Promise<CompileResult>
```

**Parameters:**

| Name | Type | Description |
|------|------|-------------|
| source | string | Contract source code |
| version | string | Solidity version |
| optimize | boolean | Enable optimization |

**Returns:** `Promise<CompileResult>`

**Example:**

```typescript
const result = await client.compile({
  source: 'pragma solidity ^0.8.0; contract MyContract {}',
  version: '0.8.19',
  optimize: true
});

console.log(result.bytecode);
```
```

## CLI Documentation

### Command Documentation

```markdown
## rechain-ide compile

Compile smart contracts.

### Usage

```bash
rechain-ide compile [options] <file>
```

### Arguments

| Name | Description |
|------|-------------|
| file | Path to the contract file |

### Options

| Option | Short | Description | Default |
|--------|-------|-------------|---------|
| --version | -v | Solidity version | latest |
| --optimize | -o | Enable optimization | false |
| --output | -O | Output directory | ./build |

### Examples

Compile a single file:
```bash
rechain-ide compile MyContract.sol
```

Compile with specific version:
```bash
rechain-ide compile MyContract.sol --version 0.8.19
```

Compile with optimization:
```bash
rechain-ide compile MyContract.sol --optimize
```
```

## Documentation Tools

### Generating API Docs

```bash
# Generate from Go code
go doc -all ./...

# Generate from OpenAPI spec
redoc-cli build openapi.yaml

# Generate TypeScript docs
typedoc --out docs/api src/
```

### OpenAPI Specification

Maintain an `openapi.yaml` file in the repository root:

```yaml
openapi: 3.0.0
info:
  title: REChain IDE API
  version: 1.0.0
paths:
  /api/v1/compile:
    post:
      summary: Compile smart contract
      requestBody:
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/CompileRequest'
      responses:
        '200':
          description: Compilation successful
```

## Style Guide

### Do's

✅ Use clear, simple language
✅ Include working examples
✅ Document all parameters
✅ Explain error cases
✅ Keep examples up-to-date
✅ Use consistent formatting

### Don'ts

❌ Leave undocumented parameters
❌ Use jargon without explanation
❌ Show broken examples
❌ Assume user knowledge
❌ Skip error documentation

## Versioning

Document which API version introduced features:

```markdown
> **Added in v1.2.0**
> This endpoint requires API version 1.2 or higher.
```

## Changelog

Keep an API changelog:

```markdown
## API Changelog

### v1.2.0 (2024-01-15)
- Added `/api/v1/projects` endpoints
- Added `debug` parameter to compile endpoint
- Deprecated `legacy` parameter (will be removed in v2.0)

### v1.1.0 (2023-12-01)
- Added GraphQL support
- Improved error messages

### v1.0.0 (2023-11-01)
- Initial API release
```

## Resources

- [OpenAPI Specification](https://swagger.io/specification/)
- [Go Doc Comments](https://go.dev/doc/comment)
- [API Blueprint](https://apiblueprint.org/)
- [GraphQL Documentation](https://graphql.org/learn/)
