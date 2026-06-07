export const CUSTOMER_PORTAL_URL =
  process.env.NEXT_PUBLIC_CUSTOMER_PORTAL_URL ?? "http://localhost:3002";

export function customerBookUrl(serviceId: string) {
  return `${CUSTOMER_PORTAL_URL}/book/${serviceId}`;
}

export function customerLoginUrl() {
  return `${CUSTOMER_PORTAL_URL}/login`;
}

export function customerRegisterUrl() {
  return `${CUSTOMER_PORTAL_URL}/register`;
}
