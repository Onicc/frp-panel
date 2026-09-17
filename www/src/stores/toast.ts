import { defineStore } from 'pinia'
export type ToastKind = 'success'|'error'
export const useToastStore = defineStore('toast', { state:()=>({items:[] as {id:number;message:string;kind:ToastKind}[]}), actions:{ show(message:string,kind:ToastKind='success'){const id=Date.now()+Math.random();this.items.push({id,message,kind});window.setTimeout(()=>this.remove(id),4000)}, remove(id:number){this.items=this.items.filter(item=>item.id!==id)} } })
