import type { Metadata } from "next";
import { QueryProvider } from "@/components/providers/query-provider";
import "./globals.css";

export const metadata: Metadata = {
  title: "Go Connect Admin",
  description: "Admin dashboard for Go Connect marketplace",
};

export default function RootLayout({
  children,
}: Readonly<{ children: React.ReactNode }>) {
  return (
    <html lang="en">
      <body className="font-sans antialiased">
        <QueryProvider>{children}</QueryProvider>
      </body>
    </html>
  );
}
