import {type ClassValue, clsx} from "clsx"
import {twMerge} from "tailwind-merge"

export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs))
}

export function getStoredUser() {
  const userStr = localStorage.getItem("user");
  if (!userStr) return null;
  const user = JSON.parse(userStr);
  const now = Math.floor(Date.now() / 1000); // seconds
  if (user.expires_at && user.expires_at < now) {
    localStorage.removeItem("user");
    return null;
  }
  return user;
}
