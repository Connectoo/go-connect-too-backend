# Payment Flow Builder

Use for Razorpay, Stripe, Cashfree, wallet, and webhooks.

Rules:
- never trust client payment status
- verify payment with provider/webhook
- use idempotency keys
- store payment states clearly
- use database transaction for booking/payment updates
- verify webhook signature
- never log payment secrets

Output:
- payment flow summary
- states added
- webhook safety checks
- failure cases handled