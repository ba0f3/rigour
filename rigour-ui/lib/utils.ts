import { clsx, type ClassValue } from "clsx"
import { twMerge } from "tailwind-merge"

export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs))
}

export function formatDate(dateString: string) {
  return new Date(dateString).toLocaleString();
}

export function formatDateShort(dateString: string) {
  return new Date(dateString).toISOString().split('T')[0];
}
