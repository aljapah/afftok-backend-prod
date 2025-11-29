export { COOKIE_NAME, ONE_YEAR_MS } from "@shared/const";

export const APP_TITLE = import.meta.env.VITE_APP_TITLE || "AffTok Admin";

export const APP_LOGO = import.meta.env.VITE_APP_LOGO || "/logo.png";

// Simplified - no OAuth required for admin panel
export const getLoginUrl = ( ) => {
  return "/login";
};
