import { defineStore } from 'pinia'
import { api } from '../api'
import type { Account } from '../api'

export const useAuthStore = defineStore('auth', {
  state: () => ({ user: null as Account | null, initialized: false }),
  actions: {
    async load() { try { this.user = await api.account() } catch { this.user = null } finally { this.initialized = true } },
    async login(username:string,password:string) { const result=await api.login(username,password); this.user=result.user },
    async register(username:string,email:string,password:string) { const result=await api.register(username,email,password); this.user=result.user },
    async logout() { await api.logout().catch(()=>undefined); this.user=null }
  }
})
