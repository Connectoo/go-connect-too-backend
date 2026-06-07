export type ApiEnvelope<T> = {
  success: boolean;
  message: string;
  data?: T;
  error?: string;
};

export type PaginatedResult<T> = {
  items: T[];
  page: number;
  limit: number;
  total: number;
};
