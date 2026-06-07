"use client";

import Link from "next/link";
import { usePathname, useRouter } from "next/navigation";
import {
  BarChart3,
  Calendar,
  CreditCard,
  FolderTree,
  Flag,
  LayoutDashboard,
  LogOut,
  MessageSquare,
  Settings,
  ShieldCheck,
  Star,
  UserCheck,
  Users,
  Wrench,
} from "lucide-react";
import { Button } from "@/components/ui/button";
import { clearAdminAuth } from "@/lib/auth";
import { cn } from "@/lib/utils";

const links = [
  { href: "/dashboard", label: "Dashboard", icon: LayoutDashboard },
  { href: "/employees", label: "Employees", icon: UserCheck },
  { href: "/kyc", label: "KYC", icon: ShieldCheck },
  { href: "/bookings", label: "Bookings", icon: Calendar },
  { href: "/users", label: "Users", icon: Users },
  { href: "/categories", label: "Categories", icon: FolderTree },
  { href: "/services", label: "Services", icon: Wrench },
  { href: "/payments", label: "Payments", icon: CreditCard },
  { href: "/subscriptions", label: "Subscriptions", icon: CreditCard },
  { href: "/subscription-plans", label: "Plans", icon: CreditCard },
  { href: "/analytics", label: "Analytics", icon: BarChart3 },
  { href: "/reviews", label: "Reviews", icon: Star },
  { href: "/reports", label: "Reports", icon: Flag },
  { href: "/support", label: "Support", icon: MessageSquare },
  { href: "/settings", label: "Settings", icon: Settings },
];

export function AdminSidebar() {
  const pathname = usePathname();
  const router = useRouter();

  return (
    <aside className="flex w-64 flex-col border-r bg-card">
      <div className="border-b p-4">
        <p className="font-semibold">Go Connect Admin</p>
        <p className="text-xs text-muted-foreground">Service marketplace</p>
      </div>
      <nav className="flex-1 space-y-1 overflow-y-auto p-3">
        {links.map(({ href, label, icon: Icon }) => (
          <Link
            key={href}
            href={href}
            className={cn(
              "flex items-center gap-2 rounded-md px-3 py-2 text-sm transition-colors",
              pathname === href || pathname.startsWith(`${href}/`)
                ? "bg-primary text-primary-foreground"
                : "text-muted-foreground hover:bg-muted hover:text-foreground",
            )}
          >
            <Icon className="h-4 w-4 shrink-0" />
            {label}
          </Link>
        ))}
      </nav>
      <div className="border-t p-3">
        <Button
          variant="ghost"
          className="w-full justify-start gap-2"
          onClick={() => {
            clearAdminAuth();
            router.push("/login");
          }}
        >
          <LogOut className="h-4 w-4" />
          Sign out
        </Button>
      </div>
    </aside>
  );
}
