# Model Drivers

## HuggingFace Driver (Inference API)

This repo includes a driver for the HuggingFace model:
- Model ID: ai-sage/GigaChat3-702B-A36B-preview
Note: The model card currently shows that this model is not deployed by any Inference Provider. citeturn1search1

### Env vars
- HF_MODEL_ID: override model id
- HF_API_URL: base URL (default https://router.huggingface.co/hf-inference/models) citeturn2search0
- HF_TOKEN: API token
- HF_TIMEOUT_MS: timeout in ms (default 8000)
- HF_FALLBACK_MODELS: comma-separated fallback model IDs
- HF_PING_TTL_MS: ping cache TTL in ms (default 15000)
- HF_PING_BACKOFF_MS: initial backoff in ms (default 1000)
- HF_PING_BACKOFF_MAX_MS: max backoff in ms (default 10000)
- HF_PING_INTERVAL_MS: background ping interval (default 60000)

## Routing constraints
- models: comma-separated driver IDs to use
- max_models: integer limit on number of drivers
- min_models: integer minimum count
- routing: latency | cost | quality | weighted | weighted_quality | quantum
- budget_ms: task timeout in ms
- max_new_tokens: model generation limit
- budget_usd: total cost cap for driver selection
- weight_cost: weight for cost in weighted routing
- weight_latency: weight for latency in weighted routing
- weight_quality: weight for quality in weighted_quality routing
- fallback_models: comma-separated fallback driver IDs
- retries: number of retries per driver
- retry_backoff_ms: backoff between retries
- quality_score: computed per result (heuristic)
- fallback_models: comma-separated fallback driver IDs

Example TaskSpec constraints:
- {"key":"models","value":"model_a,hf_gigachat3_702b_preview"}
- {"key":"max_models","value":2}

## Scoring presets
Quantum optimizer weights (env):
- balanced (default): cost=0.3, latency=0.5, quality=0.2
- latency_focus: cost=0.2, latency=0.7, quality=0.1
- quality_focus: cost=0.2, latency=0.3, quality=0.5

Agent compiler scoring weights (env):
- balanced (default): size=0.4, churn=0.3, errors=0.3
- minimal_diff: size=0.6, churn=0.3, errors=0.1
- safe_mode: size=0.3, churn=0.3, errors=0.4

## RAG Embeddings

Default embedding model:
- sentence-transformers/all-MiniLM-L6-v2 citeturn1search0
