"use client";

import Link from "next/link";
import { usePathname, useRouter } from "next/navigation";
import {
  Bell,
  Calendar,
  LayoutDashboard,
  LifeBuoy,
  LogOut,
  MapPin,
  MessageSquare,
  User,
} from "lucide-react";
import { Button } from "@/components/ui/button";
import { clearCustomerAuth } from "@/lib/auth";
import { cn } from "@/lib/utils";

const links = [
  { href: "/dashboard", label: "Dashboard", icon: LayoutDashboard },
  { href: "/bookings", label: "Bookings", icon: Calendar },
  { href: "/chat", label: "Messages", icon: MessageSquare },
  { href: "/notifications", label: "Notifications", icon: Bell },
  { href: "/profile", label: "Profile", icon: User },
  { href: "/addresses", label: "Addresses", icon: MapPin },
  { href: "/support", label: "Support", icon: LifeBuoy },
];

export function PortalSidebar() {
  const pathname = usePathname();
  const router = useRouter();

  return (
    <aside className="flex w-64 flex-col border-r bg-card">
      <div className="border-b p-4">
        <p className="font-semibold">Go Connect</p>
        <p className="text-xs text-muted-foreground">Customer portal</p>
      </div>
      <nav className="flex-1 space-y-1 p-3">
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
            clearCustomerAuth();
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
