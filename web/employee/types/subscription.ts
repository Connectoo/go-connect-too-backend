export type SubscriptionPlan = {
  id: string;
  name: string;
  price: number;
  currency: string;
  duration_days: number;
  service_limit: number;
  is_featured_allowed: boolean;
  is_priority_allowed: boolean;
  is_active: boolean;
  created_at: string;
  updated_at: string;
};

export type Subscription = {
  id: string;
  employee_id: string;
  plan_id: string;
  plan_name: string;
  status: string;
  starts_at?: string | null;
  expires_at?: string | null;
  auto_renew: boolean;
  cancelled_at?: string | null;
  cancellation_reason?: string | null;
  plan?: SubscriptionPlan | null;
  created_at: string;
  updated_at: string;
};

export type CreateOrderResponse = {
  payment_id: string;
  subscription_id: string;
  provider: string;
  provider_order_id: string;
  amount: number;
  currency: string;
  razorpay_key_id?: string;
};

export type VerifyPaymentInput = {
  payment_id: string;
  provider_order_id: string;
  provider_payment_id: string;
  signature: string;
};
