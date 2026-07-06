---
type: principle
title: Enable CSRF protection for any state-changing form
description: CodeIgniter ships CSRF middleware — leaving it off on a form is an explicit, not accidental, gap.
tags: [php, codeigniter, security, csrf]
timestamp: 2026-07-04
---

Any form or endpoint that mutates state (create/update/delete) must be
covered by CodeIgniter's CSRF protection (`csrf_protection` config enabled,
`form_open()` helper used so the hidden token field is emitted
automatically). Don't special-case an endpoint out of CSRF checking unless
it's a pure read, or is authenticated by a mechanism CSRF doesn't apply to
(e.g. a signed API token checked per-request, not cookie/session auth).
