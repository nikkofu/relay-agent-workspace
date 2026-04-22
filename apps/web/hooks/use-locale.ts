"use client"

import { useEffect, useState } from "react"
import { API_BASE_URL } from "@/lib/constants"

/**
 * Returns the user's preferred locale string (e.g. "en", "zh-CN").
 * Hydrates from GET /api/v1/me/settings on first call and falls back to
 * the browser's navigator.language.
 *
 * Stable value — only updates when settings are re-fetched.
 */
let cachedLocale: string | null = null

export function useLocale(): string {
  const [locale, setLocale] = useState<string>(
    () => cachedLocale ?? (typeof navigator !== "undefined" ? navigator.language : "en")
  )

  useEffect(() => {
    if (cachedLocale) return
    fetch(`${API_BASE_URL}/me/settings`)
      .then(r => r.ok ? r.json() : null)
      .then(data => {
        const l = data?.settings?.locale
        if (l) {
          cachedLocale = l
          setLocale(l)
        }
      })
      .catch(() => { /* silent — browser locale is fine */ })
  }, [])

  return locale
}

/**
 * Format a date string or Date object using the user's locale.
 * Uses Intl.DateTimeFormat for locale-aware output.
 */
export function formatLocaleDate(
  date: string | Date | null | undefined,
  locale: string,
  options: Intl.DateTimeFormatOptions = { year: "numeric", month: "short", day: "numeric" }
): string {
  if (!date) return ""
  try {
    return new Intl.DateTimeFormat(locale, options).format(new Date(date))
  } catch {
    return String(date)
  }
}

/**
 * Format a date as relative time using Intl.RelativeTimeFormat.
 * Falls back to a simple "X days ago" form.
 */
export function formatRelativeTime(date: string | Date | null | undefined, locale: string): string {
  if (!date) return ""
  try {
    const diff = (new Date(date).getTime() - Date.now()) / 1000
    const abs = Math.abs(diff)
    const rtf = new Intl.RelativeTimeFormat(locale, { numeric: "auto" })
    if (abs < 60) return rtf.format(Math.round(diff), "second")
    if (abs < 3600) return rtf.format(Math.round(diff / 60), "minute")
    if (abs < 86400) return rtf.format(Math.round(diff / 3600), "hour")
    if (abs < 2592000) return rtf.format(Math.round(diff / 86400), "day")
    if (abs < 31536000) return rtf.format(Math.round(diff / 2592000), "month")
    return rtf.format(Math.round(diff / 31536000), "year")
  } catch {
    return String(date)
  }
}
