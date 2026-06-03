export type TokenPair = {
  access_token: string;
  refresh_token: string;
  access_expires_at: string;
  refresh_expires_at: string;
};

export type User = {
  id: string;
  name: string;
  email: string;
  phone?: string | null;
  role: string;
  status: string;
  created_at: string;
};

export type AuthResponse = {
  user?: User;
  tokens: TokenPair;
};

export type RegisterInput = {
  name: string;
  email: string;
  phone?: string;
  password: string;
};

export type LoginInput = {
  email: string;
  password: string;
};
