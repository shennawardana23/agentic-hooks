---
type: principle
title: fetch() only rejects on network failure — check response.ok yourself
description: A 404 or 500 response is a resolved Promise, not a caught error.
tags: [javascript, fetch, error-handling]
timestamp: 2026-07-04
---

`fetch()`'s Promise only rejects for network-level failures (DNS, offline,
CORS block). An HTTP error status (4xx/5xx) still resolves normally, so
code that only wraps `fetch` in try/catch and assumes success on no
exception will silently treat error responses as success.

```js
// Bad — a 500 response is treated as success.
try {
    const res = await fetch('/api/data');
    const data = await res.json();
} catch (err) {
    console.error(err);
}

// Good — explicit status check before parsing the body.
const res = await fetch('/api/data');
if (!res.ok) {
    throw new Error(`request failed: ${res.status}`);
}
const data = await res.json();
```
