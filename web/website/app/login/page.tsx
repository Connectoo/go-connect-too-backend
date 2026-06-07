import { redirect } from "next/navigation";
import { customerLoginUrl } from "@/lib/portal-url";

export default function LoginPage() {
  redirect(customerLoginUrl());
}
