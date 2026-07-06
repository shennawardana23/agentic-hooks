---
type: principle
title: Keep controllers thin, push logic into models/services
description: CodeIgniter's MVC layering only pays off if controllers stay dispatch-only.
tags: [php, codeigniter, mvc]
timestamp: 2026-07-04
---

A CodeIgniter controller method should validate input, call into a model
(or a service class for logic spanning multiple models), and return a
view/response — not contain business logic itself. If a controller method
has more than a few lines between input validation and the response call,
that logic almost certainly belongs in the model layer where it can be
unit-tested without booting the HTTP stack.

```php
// Bad — business logic lives in the controller.
public function checkout() {
    $cart = $this->cart_model->get($this->session->userdata('cart_id'));
    $total = 0;
    foreach ($cart->items as $item) { $total += $item->price * $item->qty; }
    if ($total > 1000) { $total *= 0.9; }
    // ...
}

// Good — controller dispatches, model/service owns the rule.
public function checkout() {
    $total = $this->order_service->calculateTotal($this->session->userdata('cart_id'));
    $this->load->view('checkout', ['total' => $total]);
}
```
