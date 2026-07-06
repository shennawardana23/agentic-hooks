---
type: pattern
title: Delegate events to a stable ancestor instead of binding per-child
description: One listener on a container handles dynamically added children for free.
tags: [javascript, dom, events]
timestamp: 2026-07-04
---

When a list of similar elements (rows, list items, cards) each need the
same click behavior, attach one listener to their common parent and inspect
`event.target` (via `.closest()`), rather than attaching a listener to each
child. This avoids re-binding listeners whenever the list changes, and
avoids the memory-leak risk of removed elements retaining listeners.

```js
// Bad — must re-bind every time the list re-renders.
document.querySelectorAll('.item').forEach(el =>
    el.addEventListener('click', handleClick)
);

// Good — one listener, works for items added later too.
document.querySelector('.item-list').addEventListener('click', (event) => {
    const item = event.target.closest('.item');
    if (item) handleClick(item);
});
```
