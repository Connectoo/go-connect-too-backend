// Mirrors CategoryResponse in internal/app/spec/openapi.yaml (~line 3135).
export type Category = {
  id: string;
  name: string;
  description?: string | null;
  is_active: boolean;
  created_at: string;
  updated_at: string;
};
