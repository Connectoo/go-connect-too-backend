export type KYCRecord = {
  id: string;
  employee_id: string;
  id_proof_url: string;
  address_proof_url: string;
  status: string;
  rejection_reason?: string | null;
  reviewed_by?: string | null;
  reviewed_at?: string | null;
  created_at: string;
  updated_at: string;
};

export type SubmitKYCRequest = {
  id_proof_url: string;
  address_proof_url: string;
  id_proof_file_id?: string;
  address_proof_file_id?: string;
};

export type AdminKYCRecord = KYCRecord & {
  employee_display_name?: string | null;
  user_name: string;
  user_email: string;
};

export type KYCListResult = {
  items: AdminKYCRecord[];
  page: number;
  limit: number;
  total: number;
};
