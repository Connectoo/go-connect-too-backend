export type Payment = {
  id: string;
  employee_id: string;
  subscription_id: string;
  provider: string;
  provider_order_id: string;
  provider_payment_id?: string | null;
  amount: number;
  currency: string;
  status: string;
  created_at: string;
  updated_at: string;
};
