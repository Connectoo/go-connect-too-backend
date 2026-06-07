// Mirrors EmployeeServiceResponse in internal/app/spec/openapi.yaml (~line 3089).
export type EmployeeService = {
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

// Mirrors EmployeeServiceRequestBase / Create / Update in openapi.yaml (~line 3044).
export type EmployeeServiceRequest = {
  category_id: string;
  title: string;
  description?: string | null;
  price: number;
  duration_minutes: number;
  is_active?: boolean;
};

// Mirrors UpdateEmployeeServiceStatusRequest in openapi.yaml (~line 3082).
export type UpdateEmployeeServiceStatusRequest = {
  is_active: boolean;
};
