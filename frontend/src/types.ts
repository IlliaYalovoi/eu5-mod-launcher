export interface Mod {
  id: string
  name: string
  version: string
  tags: string[]
  description: string
  thumbnailPath: string
  dirPath: string
  enabled: boolean
  hasConflict?: boolean
  constraints?: Constraint[]
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

export type WorkspaceMode = 'load-order' | 'discover' | 'rules';

