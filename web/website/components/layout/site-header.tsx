import Link from "next/link";
import { Button } from "@/components/ui/button";
import {
  customerLoginUrl,
  customerRegisterUrl,
  employeeLoginUrl,
} from "@/lib/portal-url";

const nav = [
  { href: "/categories", label: "Categories" },
  { href: "/providers", label: "Providers" },
  { href: "/search", label: "Search" },
  { href: "/about", label: "About" },
  { href: "/contact", label: "Contact" },
];

export function SiteHeader() {
  return (
    <header className="sticky top-0 z-50 border-b bg-background/95 backdrop-blur">
      <div className="container mx-auto flex h-16 items-center justify-between px-4">
        <Link href="/" className="text-lg font-bold tracking-tight">
          Go Connect
        </Link>
        <nav className="hidden items-center gap-6 md:flex">
          {nav.map((item) => (
            <Link
              key={item.href}
              href={item.href}
              className="text-sm text-muted-foreground transition-colors hover:text-foreground"
            >
              {item.label}
            </Link>
          ))}
        </nav>
        <div className="flex items-center gap-2">
          <Button variant="ghost" asChild className="hidden sm:inline-flex">
            <a href={employeeLoginUrl()}>For providers</a>
          </Button>
          <Button variant="ghost" asChild>
            <a href={customerLoginUrl()}>Sign in</a>
          </Button>
          <Button asChild>
            <a href={customerRegisterUrl()}>Sign up</a>
          </Button>
        </div>
      </div>
    </header>
  );
}
