import { APIKey } from "./api";

/**
 * Formats an ISO date.
 * - withTime=false: "Jul 14, 2025"
 * - withTime=true:  "Jul 14, 2025, 1:45 PM" (locale-dependent)
 */
export function formatDate(
  iso: string | null | undefined,
  options: { utc?: boolean; withTime?: boolean } = {}
): string {
  if (!iso) return "";
  const d = new Date(iso);
  if (Number.isNaN(d.getTime())) return "";

  const { utc = false, withTime = false } = options;

  const base: Intl.DateTimeFormatOptions = {
    month: "short",
    day: "numeric",
    year: "numeric",
    timeZone: utc ? "UTC" : undefined,
  };

  return withTime
    ? d.toLocaleString("en-US", { ...base, hour: "2-digit", minute: "2-digit" })
    : d.toLocaleDateString("en-US", base);
}

/**
 * Returns a debounced version of the given function.
 */
export function debounce<F extends (...args: any[]) => void>(fn: F, wait = 300) {
  let t: ReturnType<typeof setTimeout> | null = null;
  return (...args: Parameters<F>) => {
    if (t) clearTimeout(t);
    t = setTimeout(() => fn(...args), wait);
  };
}

/**
 * Validates an API key name.
 * - Spaces/whitespace are disallowed (okSpaces=false if any)
 * - Allowed chars (for the non-whitespace part): A-Z a-z 0-9 . _ -
 * - Length is checked on the trimmed value (3–35)
 */
export function validateApiKeyName(raw: string) {
  const isEmpty = raw.length === 0;
  const hasWhitespace = /\s/.test(raw);
  const trimmed = raw.trim();
  const okLen = trimmed.length >= 3 && trimmed.length <= 35;
  // Ignore whitespace for the allowed-characters rule
  const cleanedForChars = trimmed.replace(/\s/g, "");
  const okChars = /^[A-Za-z0-9._-]+$/.test(cleanedForChars) || cleanedForChars.length === 0;
  const okSpaces = !hasWhitespace;

  return { raw, trimmed, isEmpty, okLen, okChars, okSpaces };
}

/**
 * Shows a Bootstrap toast notification.
 */
export function useToast(options: {
  message: string;
  title?: string;
  /** @default: 2000 */
  delay?: number;
  /** @default: true */
  autohide?: boolean;
  /** default: body-level container we create if missing */
  containerSelector?: string;
  type?: 'success' | 'error' | 'info';
  /** extra classes on .toast (e.g., "bg-success text-white") */
  className?: string;
}): HTMLElement {
  const {
    message,
    title,
    delay = 4000,
    autohide = true,
    containerSelector = "#_toast_container",
    className = "",
  } = options;

  let container = document.querySelector(containerSelector) as HTMLElement | null;
  if (!container) {
    container = document.createElement("div");
    container.id = containerSelector.replace(/^#/, "");
    container.style.position = "fixed";
    container.style.top = "1rem";
    container.style.right = "1rem";
    container.style.zIndex = "1080";
    container.style.pointerEvents = "none";
    document.body.appendChild(container);
  }

  const toast = document.createElement("div");
  toast.className = `toast ${className}`.trim();
  toast.setAttribute("role", "alert");
  toast.setAttribute("aria-live", "assertive");
  toast.setAttribute("aria-atomic", "true");
  toast.style.pointerEvents = "auto";
  const icon = options.type === "success" ? { type: "fa-check-circle", color: "text-success" } :
    options.type === "error" ? { type: "fa-times-circle", color: "text-danger" } :
      options.type === "info" ? { type: "fa-info-circle", color: "text-info" } :
        null;
  toast.innerHTML = `
    <div class="toast-body d-flex">
      ${icon ? `<i class="fa ${icon.type} fa-2x mr-3 ${icon.color}"></i>` : ""}
      <div class="flex-grow-1">
        ${title ? `
          <h6 class="m-0">${title}</h6>
        ` : ""}
        ${message}
      </div>
      <button type="button" class="ml-4 mb-1 close" data-dismiss="toast" aria-label="Close">
        <span aria-hidden="true">&times;</span>
      </button>
    </div>
  `;
  container.appendChild(toast);

  const $ = (window as any).$;
  if ($?.fn?.toast) {
    $(toast).toast({ delay, autohide }).toast("show");
    $(toast).on("hidden.bs.toast", () => toast.remove());
  } else {
    // Fallback show/hide if Bootstrap JS not present
    toast.classList.add("show");
    setTimeout(() => toast.remove(), delay);
  }
  return toast;
}

/**
 * Copies the given text to clipboard.
 */
export async function useCopyToClipboard(text: string): Promise<boolean> {
  if (!text) return false;
  try {
    if (navigator.clipboard && (window.isSecureContext ?? location.protocol === "https:")) {
      await navigator.clipboard.writeText(text);
    } else {
      const ta = document.createElement("textarea");
      ta.value = text;
      ta.setAttribute("readonly", "");
      ta.style.position = "fixed";
      ta.style.opacity = "0";
      document.body.appendChild(ta);
      ta.select();
      const ok = document.execCommand("copy");
      document.body.removeChild(ta);
      if (!ok) throw new Error("execCommand copy failed");
    }
    return true;
  } catch {
    return false;
  }
}

/**
 * Copies text to clipboard and shows a toast notification on success/failure.
 * If you dont want a toast, use directly `useCopyToClipboard`.
 * @see useCopyToClipboard
 */
export async function copyTextWithToast(text: string, options?: {
  successMessage?: string;
  failureMessage?: string;
  title?: string;
  delay?: number;
  containerSelector?: string;
}) {
  const ok = await useCopyToClipboard(text);
  if (ok) {
    useToast({
      message: options?.successMessage ?? "Copied to clipboard",
      title: options?.title ?? "Success",
      delay: options?.delay ?? 2000,
      containerSelector: options?.containerSelector,
      type: 'success',
      className: "bg-white", // keep default BS look; adjust if you want
    });
  } else {
    useToast({
      message: options?.failureMessage ?? "Copy failed",
      title: "Error",
      delay: options?.delay ?? 2500,
      containerSelector: options?.containerSelector,
      type: 'error',
      className: "bg-danger text-white",
    });
  }
  return ok;
}

/**
 * Downloads the given data as a JSON file.
 */
export function downloadJSON(data: unknown, filename: string) {
  const blob = new Blob([JSON.stringify(data, null, 2)], { type: "application/json" });
  const url = URL.createObjectURL(blob);
  const a = document.createElement("a");
  a.href = url;
  a.download = filename;
  document.body.appendChild(a);
  a.click();
  a.remove();
  URL.revokeObjectURL(url);
}

/**
 * Generates a safe filename for an API key export.
 */
export function safeFilename(base: string, ext = "json") {
  const name = (base || "api-key").trim().replace(/[^\w.-]+/g, "-").replace(/-+/g, "-");
  return `${name || "api-key"}.${ext}`;
}

/**
 * Returns a function to set validation state on an input and error container.
 */
export function useValidationState(errorContainer: HTMLElement | null) {
  return (errorMsg: string, inputEl: HTMLElement | null = null) => {
    if (errorContainer) {
      if (errorMsg.length > 0) {
        errorContainer.textContent = errorMsg;
        errorContainer.style.display = "block";
        inputEl?.focus();
        inputEl?.classList.add("is-invalid");
      } else {
        errorContainer.textContent = "";
        errorContainer.style.display = "none";
        inputEl?.classList.remove("is-invalid");
      }
    }
  }
}

/**
 * Renders API key data into the given container.
 */
export function renderKeyData(container: HTMLElement, key: APIKey) {
  const elements = container.querySelectorAll("[data-key-prop]");
  elements.forEach((el) => {
    const prop = el.getAttribute("data-key-prop") as keyof APIKey;
    if (prop in key) {
      let value = (key as any)[prop];
      if (prop === "created_at" || prop === "disabled_at" || prop === "last_used_at") {
        value = formatDate(value, { withTime: true });
      }
      el.textContent = value ?? "—";
    }
  });
}

/**
 * Clears API key data from the given container.
 */
export function clearKeyData(container: HTMLElement) {
  const elements = container.querySelectorAll("[data-key-prop]");
  elements.forEach((el) => {
    el.textContent = "";
  });
}