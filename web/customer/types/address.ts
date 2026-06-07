export type Address = {
  id: string;
  label: string;
  address_line: string;
  city: string;
  state: string;
  country: string;
  pincode: string;
  latitude?: number | null;
  longitude?: number | null;
  is_default: boolean;
  created_at: string;
  updated_at: string;
};

export type AddressInput = {
  label: string;
  address_line: string;
  city: string;
  state: string;
  country: string;
  pincode: string;
  latitude?: number | null;
  longitude?: number | null;
  is_default: boolean;
};
