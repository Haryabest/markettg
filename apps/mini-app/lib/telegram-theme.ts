type TelegramThemeParams = {
  bg_color?: string;
  text_color?: string;
  hint_color?: string;
  link_color?: string;
  button_color?: string;
  button_text_color?: string;
  secondary_bg_color?: string;
  header_bg_color?: string;
  accent_text_color?: string;
  section_bg_color?: string;
  section_header_text_color?: string;
  subtitle_text_color?: string;
  destructive_text_color?: string;
};

function setVar(root: HTMLElement, name: string, value?: string) {
  if (value) root.style.setProperty(name, value);
}

export function applyTelegramTheme(WebApp: {
  colorScheme?: string;
  themeParams?: TelegramThemeParams;
  setHeaderColor?: (color: `#${string}` | "bg_color" | "secondary_bg_color") => void;
  setBackgroundColor?: (color: `#${string}` | "bg_color" | "secondary_bg_color") => void;
}) {
  if (typeof document === "undefined") return;

  const root = document.documentElement;
  const dark = WebApp.colorScheme === "dark";
  root.classList.toggle("dark", dark);
  root.style.colorScheme = dark ? "dark" : "light";

  const tp = WebApp.themeParams ?? {};
  setVar(root, "--tg-bg-color", tp.bg_color);
  setVar(root, "--tg-text-color", tp.text_color);
  setVar(root, "--tg-hint-color", tp.hint_color);
  setVar(root, "--tg-link-color", tp.link_color);
  setVar(root, "--tg-button-color", tp.button_color);
  setVar(root, "--tg-button-text-color", tp.button_text_color);
  setVar(root, "--tg-secondary-bg-color", tp.secondary_bg_color);
  setVar(root, "--tg-header-bg-color", tp.header_bg_color);
  setVar(root, "--tg-section-bg-color", tp.section_bg_color);
  setVar(root, "--tg-subtitle-text-color", tp.subtitle_text_color);

  const bg = tp.bg_color ?? (dark ? "#0f0f0f" : "#ffffff");
  const text = tp.text_color ?? (dark ? "#fafafa" : "#1b1b1b");
  const surface = tp.secondary_bg_color ?? tp.section_bg_color ?? (dark ? "#1c1c1e" : "#f1f1f3");
  const hint = tp.hint_color ?? (dark ? "#8e8e93" : "#8a8a8e");
  const button = tp.button_color ?? (dark ? "#0a84ff" : "#007aff");
  const buttonText = tp.button_text_color ?? "#ffffff";

  setVar(root, "--background", bg);
  setVar(root, "--foreground", text);
  setVar(root, "--color-background-surface", surface);
  setVar(root, "--color-text-primary", text);
  setVar(root, "--color-text-secondary", hint);
  setVar(root, "--color-accent", button);
  setVar(root, "--color-accent-text", buttonText);
  setVar(root, "--color-border-default", dark ? "rgba(255,255,255,.12)" : "rgba(0,0,0,.08)");

  const headerColor = tp.header_bg_color ?? tp.bg_color;
  if (headerColor) WebApp.setHeaderColor?.(headerColor as `#${string}`);
  if (tp.bg_color) WebApp.setBackgroundColor?.(tp.bg_color as `#${string}`);
}

export function bindTelegramThemeListener(WebApp: {
  colorScheme?: string;
  themeParams?: TelegramThemeParams;
  setHeaderColor?: (color: `#${string}` | "bg_color" | "secondary_bg_color") => void;
  setBackgroundColor?: (color: `#${string}` | "bg_color" | "secondary_bg_color") => void;
  onEvent?: (event: "themeChanged", handler: () => void) => void;
  offEvent?: (event: "themeChanged", handler: () => void) => void;
}): () => void {
  const handler = () => applyTelegramTheme(WebApp);
  WebApp.onEvent?.("themeChanged", handler);
  return () => WebApp.offEvent?.("themeChanged", handler);
}
