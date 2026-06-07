import type { PaginatedResult } from "@/types/api";

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

export type PaymentListResult = PaginatedResult<Payment>;

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

export type SubscriptionListResult = PaginatedResult<Subscription>;

export type AdminService = {
  id: string;
  employee_id: string;
  category_id: string;
  title: string;
  description?: string | null;
  price: number;
  duration_minutes: number;
  is_active: boolean;
  created_at: string;
  updated_at: string;
};

export type ServiceListResult = PaginatedResult<AdminService>;

export type AdminReview = {
  id: string;
  booking_id: string;
  customer_id: string;
  employee_id: string;
  rating: number;
  comment?: string | null;
  review_images?: string[];
  status: string;
  created_at: string;
  updated_at: string;
};

export type ReviewListResult = PaginatedResult<AdminReview>;

export type Report = {
  id: string;
  reporter_id: string;
  reported_user_id: string;
  booking_id?: string | null;
  reason: string;
  description?: string | null;
  status: string;
  created_at: string;
  updated_at: string;
};

export type ReportListResult = PaginatedResult<Report>;

export type SupportTicket = {
  id: string;
  customer_id: string;
  subject: string;
  status: string;
  priority: string;
  created_at: string;
  updated_at: string;
};

export type SupportMessage = {
  id: string;
  ticket_id: string;
  sender_id: string;
  message: string;
  is_staff: boolean;
  created_at: string;
};

export type SupportTicketDetail = SupportTicket & {
  messages: SupportMessage[];
};

export type TicketListResult = PaginatedResult<SupportTicket>;

export type GeneralSettings = {
  site_name: string;
  support_email: string;
  maintenance_mode: boolean;
};

export type ProviderSettings = {
  razorpay_key_id?: string;
  razorpay_key_secret?: string;
  razorpay_webhook_secret?: string;
  smtp_host?: string;
  smtp_user?: string;
  smtp_pass?: string;
  smtp_from?: string;
  s3_bucket?: string;
  s3_region?: string;
  s3_access_key?: string;
  s3_secret_key?: string;
};

export type PlatformSettings = {
  general: GeneralSettings;
  provider?: ProviderSettings;
};

export type Period = {
  from: string;
  to: string;
};

export type AdminAnalyticsSummary = {
  period: Period;
  total_users: number;
  total_employees: number;
  approved_employees: number;
  active_subscriptions: number;
  monthly_recurring_revenue: number;
  booking_volume: number;
  failed_payments: number;
  churn_rate?: number | null;
};

export type AdminRevenueAnalytics = {
  period: Period;
  subscription_revenue: number;
  booking_revenue: string;
  daily_subscription_revenue: { date: string; amount: number }[];
};

export type AdminBookingsAnalytics = {
  period: Period;
  total: number;
  by_status: { status: string; count: number }[];
  daily_volume: { date: string; count: number }[];
};

export type AdminCategoriesAnalytics = {
  period: Period;
  categories: {
    category_id: string;
    category_name: string;
    booking_count: number;
  }[];
};
