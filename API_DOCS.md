# REChain Quantum-CrossAI IDE Engine API Documentation

## Overview

The REChain IDE provides a comprehensive RESTful API that allows developers to integrate with the platform programmatically. This documentation covers all available endpoints, authentication methods, and usage examples.

## API Versioning

All API requests should be prefixed with the version number:

```
https://api.rechain.ai/v1/
```

## Authentication

### API Keys

To access the API, you need an API key which can be generated in your account settings.

```bash
curl -H "Authorization: Bearer YOUR_API_KEY" \
     https://api.rechain.ai/v1/projects
```

### OAuth 2.0

For applications requiring user authorization, we support OAuth 2.0:

1. Register your application in the developer portal
2. Obtain client credentials
3. Implement the OAuth flow

## Rate Limiting

API requests are rate-limited to prevent abuse:

- 1000 requests per hour for authenticated users
- 100 requests per hour for unauthenticated users
- 10,000 requests per hour for enterprise accounts

Exceeding these limits will result in a 429 (Too Many Requests) response.

## Error Handling

All API responses follow standard HTTP status codes:

- 200: Success
- 201: Created
- 400: Bad Request
- 401: Unauthorized
- 403: Forbidden
- 404: Not Found
- 429: Too Many Requests
- 500: Internal Server Error

Error responses include a JSON object with error details:

```json
{
  "error": {
    "code": "INVALID_REQUEST",
    "message": "The request is invalid",
    "details": "Project name is required"
  }
}
```

## Projects API

### List Projects

Get a list of all projects for the authenticated user.

```
GET /projects
```

**Parameters:**
- `limit` (optional): Number of projects to return (default: 20, max: 100)
- `offset` (optional): Offset for pagination

**Response:**
```json
{
  "projects": [
    {
      "id": "proj_1234567890",
      "name": "My Project",
      "description": "A sample project",
      "created_at": "2026-01-01T00:00:00Z",
      "updated_at": "2026-01-01T00:00:00Z",
      "language": "javascript",
      "status": "active"
    }
  ],
  "total": 1,
  "limit": 20,
  "offset": 0
}
```

### Create Project

Create a new project.

```
POST /projects
```

**Request Body:**
```json
{
  "name": "My New Project",
  "description": "A description of my project",
  "language": "python",
  "template": "web-application"
}
```

**Response:**
```json
{
  "id": "proj_0987654321",
  "name": "My New Project",
  "description": "A description of my project",
  "created_at": "2026-01-01T00:00:00Z",
  "updated_at": "2026-01-01T00:00:00Z",
  "language": "python",
  "status": "active",
  "repository_url": "https://github.com/user/my-new-project"
}
```

### Get Project

Get details for a specific project.

```
GET /projects/{project_id}
```

**Response:**
```json
{
  "id": "proj_1234567890",
  "name": "My Project",
  "description": "A sample project",
  "created_at": "2026-01-01T00:00:00Z",
  "updated_at": "2026-01-01T00:00:00Z",
  "language": "javascript",
  "status": "active",
  "repository_url": "https://github.com/user/my-project",
  "ai_model": "gpt-4",
  "quantum_enabled": true,
  "web6_enabled": false
}
```

### Update Project

Update project details.

```
PUT /projects/{project_id}
```

**Request Body:**
```json
{
  "name": "Updated Project Name",
  "description": "Updated description"
}
```

### Delete Project

Delete a project.

```
DELETE /projects/{project_id}
```

## AI Agents API

### List Agents

Get a list of available AI agents.

```
GET /agents
```

**Response:**
```json
{
  "agents": [
    {
      "id": "agent_code_generator",
      "name": "Code Generator",
      "description": "Generates code from natural language descriptions",
      "capabilities": ["code_generation", "refactoring"],
      "supported_languages": ["javascript", "python", "go"]
    }
  ]
}
```

### Execute Agent

Execute an AI agent with specific parameters.

```
POST /agents/{agent_id}/execute
```

**Request Body:**
```json
{
  "project_id": "proj_1234567890",
  "parameters": {
    "prompt": "Create a function that calculates fibonacci numbers",
    "language": "python"
  }
}
```

**Response:**
```json
{
  "task_id": "task_1234567890",
  "status": "processing",
  "result": null
}
```

### Get Task Status

Get the status of an asynchronous task.

```
GET /tasks/{task_id}
```

**Response:**
```json
{
  "task_id": "task_1234567890",
  "status": "completed",
  "result": {
    "code": "def fibonacci(n):\n    if n <= 1:\n        return n\n    else:\n        return fibonacci(n-1) + fibonacci(n-2)",
    "explanation": "This function calculates fibonacci numbers recursively"
  },
  "created_at": "2026-01-01T00:00:00Z",
  "completed_at": "2026-01-01T00:00:05Z"
}
```

## Quantum API

### List Quantum Algorithms

Get a list of available quantum algorithms.

```
GET /quantum/algorithms
```

**Response:**
```json
{
  "algorithms": [
    {
      "id": "shor",
      "name": "Shor's Algorithm",
      "description": "Integer factorization algorithm",
      "complexity": "O((log N)^3)",
      "applications": ["cryptography"]
    }
  ]
}
```

### Execute Quantum Algorithm

Execute a quantum algorithm with specific parameters.

```
POST /quantum/algorithms/{algorithm_id}/execute
```

**Request Body:**
```json
{
  "parameters": {
    "number": 15
  },
  "simulation": true
}
```

**Response:**
```json
{
  "task_id": "quantum_task_1234567890",
  "status": "processing"
}
```

## Web6 API

### Create 3D Scene

Create a new 3D scene for Web6 applications.

```
POST /web6/scenes
```

**Request Body:**
```json
{
  "name": "My 3D Scene",
  "description": "A sample 3D scene"
}
```

**Response:**
```json
{
  "scene_id": "scene_1234567890",
  "name": "My 3D Scene",
  "description": "A sample 3D scene",
  "created_at": "2026-01-01T00:00:00Z",
  "updated_at": "2026-01-01T00:00:00Z"
}
```

### Add Object to Scene

Add a 3D object to a scene.

```
POST /web6/scenes/{scene_id}/objects
```

**Request Body:**
```json
{
  "type": "cube",
  "position": {
    "x": 0,
    "y": 0,
    "z": 0
  },
  "rotation": {
    "x": 0,
    "y": 0,
    "z": 0
  },
  "scale": {
    "x": 1,
    "y": 1,
    "z": 1
  }
}
```

## Collaboration API

### Invite User

Invite a user to collaborate on a project.

```
POST /projects/{project_id}/invitations
```

**Request Body:**
```json
{
  "email": "user@example.com",
  "role": "editor"
}
```

### List Collaborators

Get a list of collaborators for a project.

```
GET /projects/{project_id}/collaborators
```

## Deployment API

### Deploy Project

Deploy a project to a target environment.

```
POST /projects/{project_id}/deployments
```

**Request Body:**
```json
{
  "target": "aws",
  "environment": "production",
  "configuration": {
    "region": "us-west-2",
    "instance_type": "t3.micro"
  }
}
```

## Webhooks

The REChain IDE supports webhooks for real-time notifications about project events.

### Event Types

- `project.created`: A new project was created
- `project.updated`: A project was updated
- `project.deleted`: A project was deleted
- `deployment.completed`: A deployment was completed
- `agent.task.completed`: An AI agent task was completed

### Webhook Configuration

Configure webhooks in your account settings:

```json
{
  "url": "https://your-domain.com/webhook",
  "events": ["project.created", "deployment.completed"],
  "secret": "your-webhook-secret"
}
```

### Webhook Payload

Webhook payloads include event details and a signature for verification:

```json
{
  "event": "project.created",
  "timestamp": "2026-01-01T00:00:00Z",
  "data": {
    "project_id": "proj_1234567890",
    "name": "My Project"
  },
  "signature": "sha256=..."
}
```

## SDKs

We provide official SDKs for popular programming languages:

### JavaScript/Node.js

```javascript
const Rechain = require('@rechain/sdk');

const client = new Rechain({
  apiKey: 'YOUR_API_KEY'
});

const projects = await client.projects.list();
```

### Python

```python
from rechain import RechainClient

client = RechainClient(api_key='YOUR_API_KEY')
projects = client.projects.list()
```

### Go

```go
import "github.com/rechain/sdk-go"

client := sdk.NewClient("YOUR_API_KEY")
projects, err := client.Projects.List()
```

## Best Practices

### Error Handling

Always check the HTTP status code and handle errors appropriately:

```javascript
try {
  const response = await fetch('https://api.rechain.ai/v1/projects', {
    headers: {
      'Authorization': `Bearer ${apiKey}`
    }
  });
  
  if (!response.ok) {
    const error = await response.json();
    throw new Error(error.message);
  }
  
  const data = await response.json();
  return data;
} catch (error) {
  console.error('API request failed:', error);
}
```

### Pagination

When retrieving lists, implement pagination to handle large datasets:

```javascript
async function getAllProjects() {
  let allProjects = [];
  let offset = 0;
  const limit = 100;
  
  while (true) {
    const response = await fetch(
      `https://api.rechain.ai/v1/projects?limit=${limit}&offset=${offset}`,
      {
        headers: {
          'Authorization': `Bearer ${apiKey}`
        }
      }
    );
    
    const data = await response.json();
    
    if (data.projects.length === 0) {
      break;
    }
    
    allProjects = allProjects.concat(data.projects);
    offset += limit;
  }
  
  return allProjects;
}
```

### Rate Limiting

Implement exponential backoff for rate limiting:

```javascript
async function makeRequestWithRetry(url, options, maxRetries = 3) {
  for (let i = 0; i < maxRetries; i++) {
    try {
      const response = await fetch(url, options);
      
      if (response.status === 429) {
        const retryAfter = response.headers.get('Retry-After') || (2 ** i);
        await new Promise(resolve => setTimeout(resolve, retryAfter * 1000));
        continue;
      }
      
      return response;
    } catch (error) {
      if (i === maxRetries - 1) throw error;
      await new Promise(resolve => setTimeout(resolve, 2 ** i * 1000));
    }
  }
}
```

## Support

For API support, please contact our developer support team at developers@rechain.ai or visit our developer portal at https://developers.rechain.ai.