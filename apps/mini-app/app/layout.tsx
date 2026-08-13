import type { Metadata } from "next";
import { Geist } from "next/font/google";
import "./globals.css";
import { QueryProvider } from "@/components/providers";
import { TelegramProvider } from "@/components/telegram-provider";
import { BottomNav } from "@/components/ui";

const geist = Geist({ subsets: ["latin", "cyrillic"] });

export const metadata: Metadata = {
  title: "MarketTG",
  description: "Telegram Stars, Premium & Gifts",
};

export default function RootLayout({ children }: { children: React.ReactNode }) {
  return (
    <html lang="ru">
      <body className={`${geist.className} antialiased bg-[var(--tg-theme-bg-color,#fff)] text-[var(--tg-theme-text-color,#000)]`}>
        <QueryProvider>
          <TelegramProvider>
            <main className="mx-auto min-h-screen max-w-lg px-4 pb-24 pt-4">{children}</main>
            <BottomNav />
          </TelegramProvider>
        </QueryProvider>
      </body>
    </html>
  );
}
