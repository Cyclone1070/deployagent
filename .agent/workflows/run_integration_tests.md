---
description: Run integration tests
---

To run all integration tests (excluding paid API):
```bash
go test -v -tags=integration ./...
```

To run paid API integration tests (requires GEMINI_API_KEY):
```bash
export GEMINI_API_KEY=your_key_here
go test -v -tags="integration llm_api" ./internal/provider/gemini/...
```
