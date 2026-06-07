export type Period = {
  from: string;
  to: string;
};

export type RatingPoint = {
  period: string;
  average_rating: number;
  review_count: number;
};

export type EmployeeSummary = {
  period: Period;
  profile_views: number;
  total_bookings: number;
  completed_bookings: number;
  cancelled_bookings: number;
  average_response_time_ms: number | null;
  estimated_revenue: string;
  rating_trend: RatingPoint[];
};

export type StatusCountItem = {
  status: string;
  count: number;
};

export type DailyCountItem = {
  date: string;
  count: number;
};

export type EmployeeBookingsAnalytics = {
  period: Period;
  total: number;
  by_status: StatusCountItem[];
  daily_volume: DailyCountItem[];
};

export type EmployeeReviewsAnalytics = {
  period: Period;
  average_rating?: number | null;
  total_reviews: number;
  rating_trend: RatingPoint[];
  rating_distribution: Record<number, number>;
};
