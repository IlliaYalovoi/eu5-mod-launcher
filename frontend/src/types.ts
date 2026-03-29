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

export interface Constraint {
  type?: 'after' | 'first' | 'last'
  from?: string
  to?: string
  mod_id?: string
}

export interface LauncherCategory {
  id: string
  name: string
  mod_ids: string[]
}

export interface LauncherLayout {
  ungrouped: string[]
  categories: LauncherCategory[]
  order?: string[]
  collapsed?: Record<string, boolean>
}

