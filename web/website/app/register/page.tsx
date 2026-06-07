import { redirect } from "next/navigation";
import { customerRegisterUrl } from "@/lib/portal-url";

export default function RegisterPage() {
  redirect(customerRegisterUrl());
}
