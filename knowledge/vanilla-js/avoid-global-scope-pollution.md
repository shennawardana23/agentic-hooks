---
type: principle
title: Don't leak variables and functions into the global scope
description: Global names collide across scripts and make load order a hidden dependency.
tags: [javascript, scoping]
timestamp: 2026-07-04
---

Wrap page scripts in a module (`<script type="module">`) or at minimum an
IIFE, and avoid assigning to bare identifiers or `window.x` for anything
that isn't genuinely meant to be a cross-script public API. An accidental
global (e.g. forgetting `const`/`let` inside a non-strict script) can be
silently overwritten by a same-named variable in an unrelated script loaded
later on the same page.

```js
// Bad — pollutes global scope, order-dependent.
count = 0;
function increment() { count++; }

// Good — module scope, no leakage.
let count = 0;
export function increment() { count++; }
```
