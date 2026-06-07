export type ReviewReply = {
  id: string;
  review_id: string;
  employee_id: string;
  reply: string;
  created_at: string;
  updated_at: string;
};

export type Review = {
  id: string;
  booking_id: string;
  customer_id: string;
  employee_id: string;
  rating: number;
  comment?: string | null;
  review_images?: string[];
  status: string;
  reply?: ReviewReply | null;
  created_at: string;
  updated_at: string;
};

export type CreateReviewInput = {
  rating: number;
  comment?: string;
  review_images?: string[];
};
