import type { Metadata } from "next";
import "./globals.css";
import { AppProviders } from "@/components/app-providers";

export const metadata: Metadata = {
  title: "MarketTG",
  description: "Telegram Stars, Premium & Gifts",
};

export default function RootLayout({ children }: { children: React.ReactNode }) {
  return (
    <html lang="ru">
      <head>
        <script src="https://telegram.org/js/telegram-web-app.js" />
      </head>
      <body className="antialiased" id="__astryx-miniap">
        <AppProviders>{children}</AppProviders>
      </body>
    </html>
  );
}
