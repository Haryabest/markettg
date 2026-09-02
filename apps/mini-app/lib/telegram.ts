import type { User } from "@/lib/api";



type TelegramWebAppUser = {

  id: number;

  first_name?: string;

  last_name?: string;

  username?: string;

  photo_url?: string;

};



type TelegramWebApp = {

  initData?: string;

  initDataUnsafe?: {

    user?: TelegramWebAppUser;

  };

  ready?: () => void;

  expand?: () => void;

  colorScheme?: string;

};



declare global {

  interface Window {

    Telegram?: {

      WebApp?: TelegramWebApp;

    };

  }

}



function sleep(ms: number) {

  return new Promise((resolve) => setTimeout(resolve, ms));

}



export async function waitForTelegramWebApp(maxWaitMs = 4000): Promise<TelegramWebApp | null> {

  if (typeof window === "undefined") return null;



  const deadline = Date.now() + maxWaitMs;

  while (Date.now() < deadline) {

    const webApp = window.Telegram?.WebApp;

    if (webApp) return webApp;

    await sleep(50);

  }

  return window.Telegram?.WebApp ?? null;

}



export function parseInitDataFromHash(): string {

  if (typeof window === "undefined") return "";

  const hash = window.location.hash.replace(/^#/, "");

  if (!hash) return "";



  const params = new URLSearchParams(hash);

  const tgData = params.get("tgWebAppData");

  if (!tgData) return "";



  try {

    return decodeURIComponent(tgData).trim();

  } catch {

    return tgData.trim();

  }

}



export function parseUserFromInitData(initData: string): TelegramWebAppUser | null {

  if (!initData) return null;

  try {

    const userStr = new URLSearchParams(initData).get("user");

    if (!userStr) return null;

    const parsed = JSON.parse(userStr) as TelegramWebAppUser;

    return parsed?.id ? parsed : null;

  } catch {

    return null;

  }

}



export function readTelegramInitData(): string {

  if (typeof window === "undefined") return "";

  const fromWebApp = window.Telegram?.WebApp?.initData?.trim() || "";

  return fromWebApp || parseInitDataFromHash();

}



export function readTelegramUser(): TelegramWebAppUser | null {

  if (typeof window === "undefined") return null;

  const unsafe = window.Telegram?.WebApp?.initDataUnsafe?.user;

  if (unsafe?.id) return unsafe;



  const initData = readTelegramInitData();

  return parseUserFromInitData(initData);

}



export async function collectTelegramSession(): Promise<{

  initData: string;

  user: TelegramWebAppUser | null;

}> {

  const webApp = await waitForTelegramWebApp();

  webApp?.ready?.();

  webApp?.expand?.();



  let initData = webApp?.initData?.trim() || parseInitDataFromHash() || "";

  let user = webApp?.initDataUnsafe?.user ?? parseUserFromInitData(initData);



  if (!initData) {

    await sleep(250);

    initData = webApp?.initData?.trim() || readTelegramInitData();

    user = user ?? webApp?.initDataUnsafe?.user ?? parseUserFromInitData(initData);

  }



  return { initData, user: user?.id ? user : null };

}



export function mapTelegramUser(user: TelegramWebAppUser): User {

  return {

    id: "",

    telegram_id: user.id,

    username: user.username,

    first_name: user.first_name,

    last_name: user.last_name,

    photo_url: user.photo_url,

  };

}



export function isTelegramWebApp(): boolean {

  return Boolean(readTelegramInitData() || readTelegramUser());

}


