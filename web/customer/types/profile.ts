export type UserProfile = {
  id: string;
  name: string;
  email: string;
  phone?: string | null;
  role: string;
  status: string;
  created_at: string;
  updated_at: string;
};

export type UpdateProfileRequest = {
  name: string;
  phone?: string | null;
};

export type ChangePasswordRequest = {
  current_password: string;
  new_password: string;
};
