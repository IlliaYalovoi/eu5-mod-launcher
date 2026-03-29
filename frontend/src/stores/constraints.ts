import { ref } from 'vue'
import { defineStore } from 'pinia'
import type { Constraint } from '../types'
import { useLoadOrderStore } from './loadorder'
import {
  AddConstraint,
  AddLoadFirst,
  AddLoadLast,
  GetConstraints,
  RemoveConstraint,
  RemoveLoadFirst,
  RemoveLoadLast,
} from '../../wailsjs/go/main/App'

export const useConstraintsStore = defineStore('constraints', () => {
  const constraints = ref<Constraint[]>([])

  async function fetch(): Promise<void> {
    constraints.value = (await GetConstraints()) as Constraint[]
  }

  async function add(from: string, to: string): Promise<void> {
    await AddConstraint(from, to)
    await fetch()
  }

  async function remove(from: string, to: string): Promise<void> {
    await RemoveConstraint(from, to)
    await fetch()
  }

  async function addLoadFirst(modID: string): Promise<void> {
    await AddLoadFirst(modID)
    await fetch()
  }

  async function addLoadLast(modID: string): Promise<void> {
    await AddLoadLast(modID)
    await fetch()
  }

  async function removeLoadFirst(modID: string): Promise<void> {
    await RemoveLoadFirst(modID)
    await fetch()
  }

  async function removeLoadLast(modID: string): Promise<void> {
    await RemoveLoadLast(modID)
    await fetch()
  }

  function forMod(id: string): Constraint[] {
    if (id.indexOf('category:') === 0) {
      const loadOrderStore = useLoadOrderStore()
      const category = loadOrderStore.launcherLayout.categories.find((item) => item.id === id)
      if (!category) {
        return constraints.value.filter((c) => c.mod_id === id || c.from === id || c.to === id)
      }
      const member = new Set(category.mod_ids)
      return constraints.value.filter((c) => {
        const type = c.type ?? 'after'
        if (c.mod_id === id || c.from === id || c.to === id) {
          return true
        }
        if (type === 'first' || type === 'last') {
          return c.mod_id ? member.has(c.mod_id) : false
        }
        return (c.from ? member.has(c.from) : false) || (c.to ? member.has(c.to) : false)
      })
    }

    return constraints.value.filter((c) => {
      const type = c.type ?? 'after'
      if (type === 'first' || type === 'last') {
        return c.mod_id === id
      }
      return c.from === id || c.to === id
    })
  }

  return { constraints, fetch, add, remove, addLoadFirst, addLoadLast, removeLoadFirst, removeLoadLast, forMod }
})

