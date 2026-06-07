export type SupportTicket = {
  id: string;
  customer_id: string;
  subject: string;
  status: string;
  priority: string;
  created_at: string;
  updated_at: string;
};

export type CreateTicketInput = {
  subject: string;
  message: string;
};
