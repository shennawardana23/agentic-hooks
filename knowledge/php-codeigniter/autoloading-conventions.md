---
type: principle
title: Register shared dependencies via autoload, don't load() them per-method
description: Keeps controller/model constructors free of repeated boilerplate loads.
tags: [php, codeigniter, conventions]
timestamp: 2026-07-04
---

If a helper, library, or model is used across most controllers in the
application, add it to `application/config/autoload.php` rather than
calling `$this->load->model(...)` / `$this->load->helper(...)` at the top
of every controller method that needs it. Reserve per-method `load()` calls
for genuinely occasional dependencies used in one or two places — loading
everything globally defeats the point of autoload being a curated list.
