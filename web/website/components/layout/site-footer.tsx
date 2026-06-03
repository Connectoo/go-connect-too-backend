import Link from "next/link";

export function SiteFooter() {
  return (
    <footer className="border-t bg-muted/30">
      <div className="container mx-auto grid gap-8 px-4 py-12 md:grid-cols-3">
        <div>
          <p className="text-lg font-semibold">Go Connect</p>
          <p className="mt-2 text-sm text-muted-foreground">
            Book trusted local service professionals near you.
          </p>
        </div>
        <div className="space-y-2 text-sm">
          <p className="font-medium">Explore</p>
          <Link href="/categories" className="block text-muted-foreground hover:text-foreground">
            Categories
          </Link>
          <Link href="/providers" className="block text-muted-foreground hover:text-foreground">
            Providers
          </Link>
        </div>
        <div className="space-y-2 text-sm">
          <p className="font-medium">Legal</p>
          <Link href="/terms" className="block text-muted-foreground hover:text-foreground">
            Terms
          </Link>
          <Link href="/privacy" className="block text-muted-foreground hover:text-foreground">
            Privacy
          </Link>
        </div>
      </div>
      <div className="border-t py-4 text-center text-xs text-muted-foreground">
        © {new Date().getFullYear()} Go Connect. All rights reserved.
      </div>
    </footer>
  );
}
