import { computed } from 'vue'
import { defineStore } from 'pinia'
import type { Constraint } from '../types'
import { useLoadOrderStore } from './loadorder'
import { useSnapshotsStore } from './snapshots'
import {
  AddConstraint,
  AddLoadFirst,
  AddLoadLast,
  RemoveConstraint,
  RemoveLoadFirst,
  RemoveLoadLast,
} from '../../wailsjs/go/main/App'

export const useConstraintsStore = defineStore('constraints', () => {
  const snapshotsStore = useSnapshotsStore()

  const constraints = computed(() => snapshotsStore.activeSnapshot?.constraints || [])

  async function fetch(): Promise<void> {
    await snapshotsStore.refreshActive()
  }

  async function add(from: string, to: string): Promise<void> {
    await AddConstraint(from, to)
    await snapshotsStore.refreshActive()
  }

  async function remove(from: string, to: string): Promise<void> {
    await RemoveConstraint(from, to)
    await snapshotsStore.refreshActive()
  }

  async function addLoadFirst(modID: string): Promise<void> {
    await AddLoadFirst(modID)
    await snapshotsStore.refreshActive()
  }

  async function addLoadLast(modID: string): Promise<void> {
    await AddLoadLast(modID)
    await snapshotsStore.refreshActive()
  }

  async function removeLoadFirst(modID: string): Promise<void> {
    await RemoveLoadFirst(modID)
    await snapshotsStore.refreshActive()
  }

  async function removeLoadLast(modID: string): Promise<void> {
    await RemoveLoadLast(modID)
    await snapshotsStore.refreshActive()
  }

  function forMod(id: string): Constraint[] {
    if (id.indexOf('category:') === 0) {
      const loadOrderStore = useLoadOrderStore()
      const category = loadOrderStore.launcherLayout.categories.find((item) => item.id === id)
      if (!category) {
        return constraints.value.filter((c) => c.modId === id || c.from === id || c.to === id)
      }
      const member = new Set(category.modIds)
      return constraints.value.filter((c) => {
        const type = c.type ?? 'after'
        if (c.modId === id || c.from === id || c.to === id) {
          return true
        }
        if (type === 'first' || type === 'last') {
          return c.modId ? member.has(c.modId) : false
        }
        return (c.from ? member.has(c.from) : false) || (c.to ? member.has(c.to) : false)
      })
    }

    return constraints.value.filter((c) => {
      const type = c.type ?? 'after'
      if (type === 'first' || type === 'last') {
        return c.modId === id
      }
      return c.from === id || c.to === id
    })
  }

  return { constraints, fetch, add, remove, addLoadFirst, addLoadLast, removeLoadFirst, removeLoadLast, forMod }
})
