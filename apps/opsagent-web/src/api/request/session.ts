/**
 * 会话凭证与登录跳转（本应用无本地登录页，凭共享平台 cookie 会话）。
 *
 * ecmdb-token-key 是 eiam 签发、e-cam-service / 控制台 / 本应用共用的平台凭证；
 * 请求层读取 cookie → Bearer 透传，401 统一跳运维平台登录页（/console/login）。
 * Hard Rule：不清共享 cookie（清除是 eiam logout 的职责，局部故障不得扩散为全平台掉线）。
 */

/** 会话 cookie 键（eiam 签发，与 haven-console src/api/request/eiam.ts 同源） */
export const SESSION_COOKIE_KEY = "ecmdb-token-key";

/** 从共享 cookie 读取会话 token（JWT 为 URL-safe base64，无需容错解码） */
export function getSessionToken(): string | undefined {
  if (typeof document === "undefined") return undefined;
  const match = document.cookie.match(
    new RegExp(`(?:^|;\\s*)${SESSION_COOKIE_KEY}=([^;]*)`),
  );
  return match?.[1] ? decodeURIComponent(match[1]) : undefined;
}

/** 会话 cookie 是否存在（同步、不发请求） */
export function hasSessionCookie(): boolean {
  return getSessionToken() !== undefined;
}

/**
 * 登录页目标：
 * - dev（vite 5175 直连）：跨源跳 http://127.0.0.1:8888/console/login（nginx 同域形态）
 * - prod（nginx /opsagent/ 托管）：同源 /console/login
 */
export function loginTarget(): string {
  return import.meta.env.DEV
    ? "http://127.0.0.1:8888/console/login"
    : "/console/login";
}

/** 防止多个并发 401 请求重复跳转（模块级单次置位，不复位） */
let isRedirectingToLogin = false;

/** 统一跳转登录页（带 redirect 回跳；Hard Rule：不清共享 cookie） */
export function redirectToLogin(): void {
  if (isRedirectingToLogin) return;
  isRedirectingToLogin = true;
  const currentUrl = window.location.href;
  window.location.href = `${loginTarget()}?redirect=${encodeURIComponent(currentUrl)}`;
}