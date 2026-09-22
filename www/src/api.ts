export type Account = { username:string; email:string; role:string }
export type VersionInfo = { gitVersion:string; gitCommit:string; gitBranch?:string; buildDate:string; platform:string }
export type UpdateOperation = { id:string; kind:'master'|'server'; targetId:string; targetCommit:string; version:string; state:'queued'|'downloading'|'staged'|'restarting'|'verifying'|'succeeded'|'failed'|'rolled_back'; error?:string; startedAt:string; updatedAt:string; completedAt?:string }
export type ReleaseSnapshot = { channel:'legacy'|'stable'|'unknown'; currentVersion:string; currentCommit:string; latestVersion?:string; latestCommit?:string; releaseUrl?:string; publishedAt?:string; available:boolean|null; checkedAt:string; error?:string }
export type Client = { id:string; comment:string; configurationState:'configured'|'unconfigured'; status:'pending'|'online'|'offline'|'error'|'disabled'; enabled:boolean; lastSeenAt?:string; enrolledAt?:string; tunnelCount:number; locationIpOverride?:string; version?:VersionInfo; versionAt?:string }
export type Server = { id:string; address:string; bindPort:number; serverApiPort:number; comment:string; configurationState:'configured'|'unconfigured'; status:'pending'|'online'|'offline'|'error'|'disabled'; lastSeenAt?:string; enrolledAt?:string; tunnelCount:number; version?:VersionInfo; versionAt?:string; autoUpdate:boolean; updateZone:string; updateStart:string; updateEnd:string; updateOperation?:UpdateOperation }
export type Tunnel = { id:string; name:string; clientId:string; serverId:string; serverAddress?:string; type:'tcp'|'udp'; localHost:string; localPort:number; remotePort:number; enabled:boolean; status:string; lastError?:string; updatedAt:string }
export type TopologyNode = { id:string; kind:'client'|'server'; label:string; comment?:string; address?:string; status:string; configurationState:string; enabled:boolean; locationIp?:string; locationSource?:'manual'|'agent_probe'|'observed'|'configured'; observedIp?:string; reportedIp?:string; reportedAt?:string; lastSeenAt?:string; tunnelCount:number }
export type TopologyLink = { id:string; name:string; sourceClientId:string; targetServerId:string; type:'tcp'|'udp'; remotePort:number; enabled:boolean; status:string; lastError?:string }
export type TopologyResponse = { nodes:TopologyNode[]; links:TopologyLink[]; locatedCount:number; totalCount:number; generatedAt:string }
export type Page<T> = { items:T[]; total:number; page:number; pageSize:number }
export type Enrollment = { clientId?:string; serverId?:string; token:string; expiresAt:string; apiUrl:string; rpcUrl:string; installCommand?:string; installCommands?:Record<'linux'|'darwin'|'windows',string>; composeYaml?:string }
export class APIError extends Error { constructor(message:string, readonly status:number, readonly problem?:any){super(message);this.name='APIError'} }

async function request<T>(path:string, init:RequestInit={}):Promise<T>{
  const headers=new Headers(init.headers); headers.set('Accept','application/json'); if(init.body!==undefined)headers.set('Content-Type','application/json')
  const response=await fetch(path,{...init,headers,credentials:'same-origin'}); let payload:any={}; if(response.status!==204)payload=await response.json().catch(()=>({}))
  if(!response.ok){if(response.status===401)window.dispatchEvent(new Event('frp-panel:unauthorized'));const message=payload.detail||payload.title||payload.msg||`Request failed (${response.status})`;throw new APIError(message,response.status,payload)}
  return payload as T
}
const query=(params:Record<string,string|number|undefined>)=>{const q=new URLSearchParams();Object.entries(params).forEach(([k,v])=>{if(v!==undefined&&v!=='')q.set(k,String(v))});const value=q.toString();return value?`?${value}`:''}
export const api={
  login:(username:string,password:string)=>request<{user:Account}>('/api/v2/auth/login',{method:'POST',body:JSON.stringify({username,password})}),
  register:(username:string,email:string,password:string)=>request<{user:Account}>('/api/v2/auth/register',{method:'POST',body:JSON.stringify({username,email,password})}),
  bootstrapStatus:()=>request<{registrationEnabled:boolean;ownerExists:boolean;canCreateOwner:boolean}>('/api/v2/bootstrap-status'),
  logout:()=>request<void>('/api/v2/auth/logout',{method:'POST'}),
  account:async()=>{const response=await request<{user?:Account;username?:string;email?:string;role?:string}>('/api/v2/account');return response.user||{username:response.username||'',email:response.email||'',role:response.role||''}},
  changePassword:(currentPassword:string,newPassword:string)=>request<{reauthenticate:boolean}>('/api/v2/account/password',{method:'POST',body:JSON.stringify({currentPassword,newPassword})}),
  overview:()=>request<{clients:number;servers:number;tunnels:number}>('/api/v2/overview'),
  topology:()=>request<TopologyResponse>('/api/v2/topology'),
  release:(channel?:'stable',force=false)=>request<{release:ReleaseSnapshot;supported:boolean;operation?:UpdateOperation}>(`/api/v2/updates/release${query({channel,force:force?'true':undefined})}`),
  updateMaster:()=>request<{operationId:string}>('/api/v2/updates/master',{method:'POST'}),
  updateOperation:(id:string)=>request<{operation:UpdateOperation}>(`/api/v2/updates/operations/${encodeURIComponent(id)}`),
  clients:(params:Record<string,string|number|undefined>={})=>request<Page<Client>>(`/api/v2/clients${query(params)}`),
  createClient:(body:{clientId:string;comment?:string})=>request<{client:Client;enrollment:Enrollment}>('/api/v2/clients',{method:'POST',body:JSON.stringify(body)}),
  updateClient:(id:string,body:Record<string,unknown>)=>request<{client:Client}>(`/api/v2/clients/${encodeURIComponent(id)}`,{method:'PATCH',body:JSON.stringify(body)}),
  deleteClient:(id:string)=>request<void>(`/api/v2/clients/${encodeURIComponent(id)}`,{method:'DELETE'}),
  rotateClient:(id:string,acknowledgeDisruption:boolean)=>request<{client:Client;enrollment:Enrollment}>(`/api/v2/clients/${encodeURIComponent(id)}/enrollment`,{method:'POST',body:JSON.stringify({acknowledgeDisruption})}),
  servers:(params:Record<string,string|number|undefined>={})=>request<Page<Server>>(`/api/v2/servers${query(params)}`),
  createServer:(body:{serverId:string;address:string;bindPort:number;serverApiPort:number;comment?:string})=>request<{server:Server;enrollment:Enrollment}>('/api/v2/servers',{method:'POST',body:JSON.stringify(body)}),
  updateServer:(id:string,body:Record<string,unknown>)=>request<{server:Server}>(`/api/v2/servers/${encodeURIComponent(id)}`,{method:'PATCH',body:JSON.stringify(body)}),
  deleteServer:(id:string)=>request<void>(`/api/v2/servers/${encodeURIComponent(id)}`,{method:'DELETE'}),
  rotateServer:(id:string,acknowledgeDisruption:boolean)=>request<{server:Server;enrollment:Enrollment}>(`/api/v2/servers/${encodeURIComponent(id)}/enrollment`,{method:'POST',body:JSON.stringify({acknowledgeDisruption})}),
  setServerUpdatePolicy:(id:string,body:{enabled:boolean;timeZone:string;windowStart:string;windowEnd:string})=>request<void>(`/api/v2/servers/${encodeURIComponent(id)}/update-policy`,{method:'PATCH',body:JSON.stringify(body)}),
  updateServerBinary:(id:string)=>request<{operationId:string}>(`/api/v2/servers/${encodeURIComponent(id)}/updates`,{method:'POST'}),
  tunnels:(params:Record<string,string|number|undefined>={})=>request<Page<Tunnel>>(`/api/v2/tunnels${query(params)}`),
  createTunnel:(body:Record<string,unknown>)=>request<{tunnel:Tunnel}>('/api/v2/tunnels',{method:'POST',body:JSON.stringify(body)}),
  updateTunnel:(id:string,body:Record<string,unknown>)=>request<{tunnel:Tunnel}>(`/api/v2/tunnels/${encodeURIComponent(id)}`,{method:'PATCH',body:JSON.stringify(body)}),
  deleteTunnel:(id:string)=>request<void>(`/api/v2/tunnels/${encodeURIComponent(id)}`,{method:'DELETE'})
}
