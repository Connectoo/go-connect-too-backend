import { apiRequest } from "@/lib/api-client";
import { getAccessToken } from "@/lib/auth";
import type { Address, AddressInput } from "@/types/address";
import type {
  ChangePasswordRequest,
  UpdateProfileRequest,
  UserProfile,
} from "@/types/profile";

function authOptions(extra?: RequestInit) {
  return {
    token: getAccessToken(),
    ...extra,
  };
}

export function fetchProfile() {
  return apiRequest<UserProfile>("/users/me", authOptions());
}

export function updateProfile(body: UpdateProfileRequest) {
  return apiRequest<UserProfile>("/users/me", {
    ...authOptions({ method: "PUT", body: JSON.stringify(body) }),
  });
}

export function deactivateAccount() {
  return apiRequest<UserProfile>("/users/me/deactivate", {
    ...authOptions({ method: "PATCH" }),
  });
}

export function changePassword(body: ChangePasswordRequest) {
  return apiRequest("/auth/change-password", {
    ...authOptions({ method: "POST", body: JSON.stringify(body) }),
  });
}

export function resendVerificationEmail() {
  return apiRequest("/auth/resend-verification", {
    ...authOptions({ method: "POST" }),
  });
}

export function fetchAddresses() {
  return apiRequest<Address[]>("/users/addresses", authOptions());
}

export function createAddress(body: AddressInput) {
  return apiRequest<Address>("/users/addresses", {
    ...authOptions({ method: "POST", body: JSON.stringify(body) }),
  });
}

export function updateAddress(id: string, body: AddressInput) {
  return apiRequest<Address>(`/users/addresses/${id}`, {
    ...authOptions({ method: "PUT", body: JSON.stringify(body) }),
  });
}

export function deleteAddress(id: string) {
  return apiRequest(`/users/addresses/${id}`, {
    ...authOptions({ method: "DELETE" }),
  });
}
