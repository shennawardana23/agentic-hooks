---
type: principle
title: Use CodeIgniter's Query Builder over raw SQL string concatenation
description: The Query Builder auto-escapes identifiers and bound values, closing the most common SQL-injection path.
tags: [php, codeigniter, sql, security]
timestamp: 2026-07-04
---

Prefer `$this->db->where(...)`, `->get()`, `->insert()` etc over
`$this->db->query("SELECT * FROM users WHERE id = " . $id)`. If a query is
too complex for the builder, use `query()` with bound parameters (`?`
placeholders and a values array) rather than string interpolation — never
concatenate user input directly into SQL text.

```php
// Bad — direct injection vector.
$this->db->query("SELECT * FROM users WHERE email = '$email'");

// Good — Query Builder escapes automatically.
$this->db->where('email', $email)->get('users');

// Good — raw query, but parameterized.
$this->db->query('SELECT * FROM users WHERE email = ?', [$email]);
```
