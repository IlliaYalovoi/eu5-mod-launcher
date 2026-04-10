export interface Mod {
  ID: string
  Name: string
  Version: string
  Tags: string[]
  Description: string
  ThumbnailPath: string
  DirPath: string
  Enabled: boolean
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

