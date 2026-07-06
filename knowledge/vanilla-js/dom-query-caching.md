---
type: pattern
title: Cache repeated DOM queries instead of re-querying in hot paths
description: querySelector walks the tree every call — cheap once, expensive in a loop or handler fired often.
tags: [javascript, dom, performance]
timestamp: 2026-07-04
---

If the same element is queried more than once inside a loop or a frequently
invoked event handler (scroll, input, resize), query it once outside the
hot path and reuse the reference, rather than calling
`document.querySelector` again on every invocation.

```js
// Bad — re-queries the DOM on every scroll event.
window.addEventListener('scroll', () => {
    document.querySelector('.progress-bar').style.width = `${pct}%`;
});

// Good — queried once, reused.
const progressBar = document.querySelector('.progress-bar');
window.addEventListener('scroll', () => {
    progressBar.style.width = `${pct}%`;
});
```
