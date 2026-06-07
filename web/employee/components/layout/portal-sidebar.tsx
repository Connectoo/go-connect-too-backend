"use client";

import Link from "next/link";
import { usePathname, useRouter } from "next/navigation";
import {
  BarChart3,
  Bell,
  CalendarClock,
  CreditCard,
  Receipt,
  LayoutDashboard,
  LogOut,
  MessageSquare,
  ShieldCheck,
  Star,
  User,
  Wrench,
} from "lucide-react";
import { Button } from "@/components/ui/button";
import { clearEmployeeAuth } from "@/lib/auth";
import { cn } from "@/lib/utils";

const links = [
  { href: "/dashboard", label: "Dashboard", icon: LayoutDashboard },
  { href: "/bookings", label: "Bookings", icon: CalendarClock },
  { href: "/services", label: "Services", icon: Wrench },
  { href: "/availability", label: "Availability", icon: CalendarClock },
  { href: "/reviews", label: "Reviews", icon: Star },
  { href: "/subscription", label: "Subscription", icon: CreditCard },
  { href: "/payments", label: "Payments", icon: Receipt },
  { href: "/analytics", label: "Analytics", icon: BarChart3 },
  { href: "/chat", label: "Messages", icon: MessageSquare },
  { href: "/notifications", label: "Notifications", icon: Bell },
  { href: "/profile", label: "Profile", icon: User },
  { href: "/kyc", label: "KYC", icon: ShieldCheck },
];

export function PortalSidebar() {
  const pathname = usePathname();
  const router = useRouter();

  return (
    <aside className="flex w-64 flex-col border-r bg-card">
      <div className="border-b p-4">
        <p className="font-semibold">Go Connect Pro</p>
        <p className="text-xs text-muted-foreground">Provider portal</p>
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
            <Icon className="h-4 w-4" />
            {label}
          </Link>
        ))}
      </nav>
      <div className="border-t p-3">
        <Button
          variant="ghost"
          className="w-full justify-start gap-2"
          onClick={() => {
            clearEmployeeAuth();
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
