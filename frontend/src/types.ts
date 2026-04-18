export interface Mod {
  ID: string
  Name: string
  Version: string
  SupportedVersion: string
  Tags: string[]
  Description: string
  ThumbnailPath: string
  DirPath: string
  Enabled: boolean
  IsCompatible: boolean
}

export interface WorkshopItem {
  itemId: string
  title: string
  description: string
  previewUrl: string
}

export interface Constraint {
  type?: 'after' | 'first' | 'last'
  from?: string
  to?: string
  modId?: string
}

export interface LauncherCategory {
  id: string
  name: string
  modIds: string[]
}

export interface LauncherLayout {
  ungrouped: string[]
  categories: LauncherCategory[]
  order?: string[]
  collapsed?: Record<string, boolean>
}

export interface GameSettingsData {
  modsDir?: string
  gameExe?: string
  gameVersionOverride?: string
}

export interface SnapshotMeta {
  revision: number
  fetchedAt: number
  stale: boolean
}

export interface SnapshotModsDirStatus {
  effectiveDir: string
  autoDetectedDir: string
  customDir: string
  usingCustomDir: boolean
  autoDetectedExists: boolean
  effectiveExists: boolean
}

export interface SnapshotSettings {
  modsDirStatus: SnapshotModsDirStatus
  gameExe: string
  autoDetectedGameExe: string
  configPath: string
  gameVersion: string
  gameVersionOverride: string
  availableGames: string[]
}

export interface GameSnapshot {
  gameID: string
  mods: Mod[]
  loadOrder: string[]
  launcherLayout: LauncherLayout
  constraints: Constraint[]
  playsetNames: string[]
  gameActivePlaysetIndex: number
  launcherActivePlaysetIndex: number
  settings: SnapshotSettings
  meta: SnapshotMeta
}

