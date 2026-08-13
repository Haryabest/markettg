import { type ClassValue, clsx } from "clsx";
import { twMerge } from "tailwind-merge";

export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs));
}

export function formatPrice(kopecks: number): string {
  return `${(kopecks / 100).toFixed(0)} ₽`;
}
