import { create } from "zustand"
import { API_BASE_URL } from "@/lib/constants"
import {
  buildBusinessAppCacheKey,
  buildBusinessAppQueryString,
  normalizeBusinessApp,
  normalizeBusinessAppDataResponse,
  normalizeBusinessAppsResponse,
  normalizeBusinessAppStatsResponse,
} from "@/lib/business-apps"
import type {
  BusinessApp,
  BusinessAppDataResponse,
  BusinessAppQueryState,
  BusinessAppStatsResponse,
} from "@/types"

interface BusinessAppState {
  apps: BusinessApp[]
  appsById: Record<string, BusinessApp | null>
  dataByKey: Record<string, BusinessAppDataResponse>
  statsByKey: Record<string, BusinessAppStatsResponse>
  isLoadingApps: boolean
  isLoadingApp: Record<string, boolean>
  isLoadingData: Record<string, boolean>
  isLoadingStats: Record<string, boolean>
  appsError: string | null
  appErrors: Record<string, string | null>
  dataErrors: Record<string, string | null>
  statsErrors: Record<string, string | null>
  fetchApps: () => Promise<BusinessApp[]>
  fetchApp: (appKey: string) => Promise<BusinessApp | null>
  fetchAppData: (appKey: string, query: BusinessAppQueryState) => Promise<BusinessAppDataResponse | null>
  fetchAppStats: (appKey: string, query: Partial<BusinessAppQueryState>) => Promise<BusinessAppStatsResponse | null>
}

export const useBusinessAppStore = create<BusinessAppState>((set, get) => ({
  apps: [],
  appsById: {},
  dataByKey: {},
  statsByKey: {},
  isLoadingApps: false,
  isLoadingApp: {},
  isLoadingData: {},
  isLoadingStats: {},
  appsError: null,
  appErrors: {},
  dataErrors: {},
  statsErrors: {},

  fetchApps: async () => {
    set({ isLoadingApps: true, appsError: null })
    try {
      const response = await fetch(`${API_BASE_URL}/apps`)
      if (!response.ok) {
        throw new Error("Failed to fetch apps")
      }
      const data = await response.json()
      const apps = normalizeBusinessAppsResponse(data)
      set(state => ({
        apps,
        appsById: {
          ...state.appsById,
          ...Object.fromEntries(apps.map(app => [app.key, app])),
        },
        isLoadingApps: false,
        appsError: null,
      }))
      return apps
    } catch (error) {
      const message = error instanceof Error ? error.message : "Failed to fetch apps"
      console.error("Failed to fetch apps:", error)
      set({ isLoadingApps: false, appsError: message })
      return []
    }
  },

  fetchApp: async (appKey) => {
    set(state => ({
      isLoadingApp: { ...state.isLoadingApp, [appKey]: true },
      appErrors: { ...state.appErrors, [appKey]: null },
    }))
    try {
      const response = await fetch(`${API_BASE_URL}/apps/${encodeURIComponent(appKey)}`)
      if (!response.ok) {
        throw new Error(`Failed to fetch ${appKey} app metadata`)
      }
      const data = await response.json()
      const app = normalizeBusinessApp((data as { app?: unknown }).app)
      if (!app) {
        throw new Error(`Invalid ${appKey} app metadata`)
      }
      set(state => ({
        apps: state.apps.some(existing => existing.key === app.key)
          ? state.apps.map(existing => (existing.key === app.key ? app : existing))
          : [...state.apps, app],
        appsById: { ...state.appsById, [app.key]: app },
        isLoadingApp: { ...state.isLoadingApp, [appKey]: false },
        appErrors: { ...state.appErrors, [appKey]: null },
      }))
      return app
    } catch (error) {
      const message = error instanceof Error ? error.message : `Failed to fetch ${appKey} app metadata`
      console.error(`Failed to fetch ${appKey} app metadata:`, error)
      set(state => ({
        appsById: { ...state.appsById, [appKey]: null },
        isLoadingApp: { ...state.isLoadingApp, [appKey]: false },
        appErrors: { ...state.appErrors, [appKey]: message },
      }))
      return null
    }
  },

  fetchAppData: async (appKey, query) => {
    const cacheKey = buildBusinessAppCacheKey(appKey, query)
    set(state => ({
      isLoadingData: { ...state.isLoadingData, [cacheKey]: true },
      dataErrors: { ...state.dataErrors, [cacheKey]: null },
    }))
    try {
      const queryString = buildBusinessAppQueryString(query)
      const response = await fetch(`${API_BASE_URL}/apps/${encodeURIComponent(appKey)}/data${queryString ? `?${queryString}` : ""}`)
      if (!response.ok) {
        throw new Error(`Failed to fetch ${appKey} app data`)
      }
      const data = await response.json()
      const fallbackApp = get().appsById[appKey] ?? null
      const normalized = normalizeBusinessAppDataResponse(data, query.mode, fallbackApp)
      set(state => ({
        dataByKey: { ...state.dataByKey, [cacheKey]: normalized },
        isLoadingData: { ...state.isLoadingData, [cacheKey]: false },
        dataErrors: { ...state.dataErrors, [cacheKey]: null },
      }))
      return normalized
    } catch (error) {
      const message = error instanceof Error ? error.message : `Failed to fetch ${appKey} app data`
      console.error(`Failed to fetch ${appKey} app data:`, error)
      set(state => ({
        isLoadingData: { ...state.isLoadingData, [cacheKey]: false },
        dataErrors: { ...state.dataErrors, [cacheKey]: message },
      }))
      return null
    }
  },

  fetchAppStats: async (appKey, query) => {
    const cacheKey = buildBusinessAppCacheKey(appKey, query)
    set(state => ({
      isLoadingStats: { ...state.isLoadingStats, [cacheKey]: true },
      statsErrors: { ...state.statsErrors, [cacheKey]: null },
    }))
    try {
      const queryString = buildBusinessAppQueryString(query, { includeCursor: false, includeLimit: false })
      const response = await fetch(`${API_BASE_URL}/apps/${encodeURIComponent(appKey)}/stats${queryString ? `?${queryString}` : ""}`)
      if (!response.ok) {
        throw new Error(`Failed to fetch ${appKey} app stats`)
      }
      const data = await response.json()
      const fallbackApp = get().appsById[appKey] ?? null
      const normalized = normalizeBusinessAppStatsResponse(data, fallbackApp)
      set(state => ({
        statsByKey: { ...state.statsByKey, [cacheKey]: normalized },
        isLoadingStats: { ...state.isLoadingStats, [cacheKey]: false },
        statsErrors: { ...state.statsErrors, [cacheKey]: null },
      }))
      return normalized
    } catch (error) {
      const message = error instanceof Error ? error.message : `Failed to fetch ${appKey} app stats`
      console.error(`Failed to fetch ${appKey} app stats:`, error)
      set(state => ({
        isLoadingStats: { ...state.isLoadingStats, [cacheKey]: false },
        statsErrors: { ...state.statsErrors, [cacheKey]: message },
      }))
      return null
    }
  },
}))
