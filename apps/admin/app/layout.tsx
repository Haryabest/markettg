import type { Metadata } from "next";
import { Geist } from "next/font/google";
import "./globals.css";
import { AdminShell } from "@/components/admin-shell";

const geist = Geist({ subsets: ["latin", "cyrillic"] });

export const metadata: Metadata = {
  title: "MarketTG Admin",
};

export default function RootLayout({ children }: { children: React.ReactNode }) {
  return (
    <html lang="ru">
      <body className={`${geist.className} antialiased`}>
        <AdminShell>{children}</AdminShell>
      </body>
    </html>
  );
}
