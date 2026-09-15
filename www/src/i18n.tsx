import { createContext, type PropsWithChildren, useContext, useMemo, useState } from 'react'

const messages = {
  zh: {
    overview: '概览', nodes: '节点', servers: '服务端', tunnels: '隧道', networks: '网络', functions: '函数',
    audit: '审计日志', users: '用户与权限', settings: '设置', signOut: '退出登录', online: '在线', offline: '离线',
    createNode: '添加节点', createServer: '创建服务端', refresh: '刷新', cancel: '取消', confirm: '确认', saving: '正在保存…',
    nodeId: '节点 ID', serverId: '服务端 ID', serverIp: '服务端 IP', installAgent: '安装 Agent', copy: '复制命令', copied: '已复制',
	linux: 'Linux', macos: 'macOS', windows: 'Windows', noData: '暂无数据', login: '登录', username: '用户名或邮箱', password: '密码', email: '邮箱',
	bootstrapOwner: '首次部署？创建所有者', createOwner: '创建所有者并登录', backToLogin: '返回登录',
    welcome: '安全地管理你的 FRP 网络', controlPlane: '控制面', capabilityHint: '高权限能力默认关闭，并按节点系统能力显示。',
    comingSoon: '此模块已纳入 v2 控制面；当前节点不支持的操作会被明确禁用。', error: '操作失败', submitSuccess: '设置已保存',
    status: '状态', lifecycle: '生命周期', ephemeral: '临时', managed: '受管', loading: '加载中…', bindPort: 'FRPS 绑定端口',
    deployServer: '部署 FRPS', createDeployment: '创建并生成部署文件', copyCompose: '复制 compose.yaml',
    serverTopologyHint: '每台公网服务器独立运行一个受管 FRPS；Master 不承载代理流量。',
    serverTokenHint: '将此 compose.yaml 保存到目标服务器。一次性令牌 10 分钟有效，首次启动后长期凭据只保存在数据卷中。',
    serverCreateHint: '公网地址供 FRPC 连接；目标服务器需放行绑定端口和隧道使用的 remote port。',
    nodeTopologyHint: '每个 Agent 可由后台分配一个或多个 FRPS 链路。', agentTokenHint: '注册令牌 10 分钟有效且只能使用一次。',
    nodeCreateHint: '长期凭据只会写入目标系统受保护的 Agent 配置文件。', frpsRoutes: 'FRPS 链路', addRoute: '分配链路',
    removeRoute: '移除链路', unassigned: '未分配', createServerFirst: '请先创建 FRPS', selectServer: '选择 FRPS', add: '添加',
  },
  en: {
    overview: 'Overview', nodes: 'Nodes', servers: 'Servers', tunnels: 'Tunnels', networks: 'Networks', functions: 'Functions',
    audit: 'Audit log', users: 'Users & access', settings: 'Settings', signOut: 'Sign out', online: 'Online', offline: 'Offline',
    createNode: 'Add node', createServer: 'Create server', refresh: 'Refresh', cancel: 'Cancel', confirm: 'Confirm', saving: 'Saving…',
    nodeId: 'Node ID', serverId: 'Server ID', serverIp: 'Server IP', installAgent: 'Install agent', copy: 'Copy command', copied: 'Copied',
	linux: 'Linux', macos: 'macOS', windows: 'Windows', noData: 'No data yet', login: 'Sign in', username: 'Username or email', password: 'Password', email: 'Email',
	bootstrapOwner: 'First run? Create owner', createOwner: 'Create owner and sign in', backToLogin: 'Back to sign in',
    welcome: 'Manage your FRP network securely', controlPlane: 'Control plane', capabilityHint: 'Privileged features are off by default and gated by node capabilities.',
    comingSoon: 'This module is part of the v2 control plane. Unsupported node actions are explicitly disabled.', error: 'Operation failed', submitSuccess: 'Settings saved',
    status: 'Status', lifecycle: 'Lifecycle', ephemeral: 'Ephemeral', managed: 'Managed', loading: 'Loading…', bindPort: 'FRPS bind port',
    deployServer: 'Deploy FRPS', createDeployment: 'Create deployment', copyCompose: 'Copy compose.yaml',
    serverTopologyHint: 'Each public server runs an independently managed FRPS; Master carries no proxy traffic.',
    serverTokenHint: 'Save this compose.yaml on the target server. The one-use token expires in 10 minutes; long-lived credentials remain only in its data volume.',
    serverCreateHint: 'FRPC connects to this public address. Allow the bind port and each tunnel remote port on the target server.',
    nodeTopologyHint: 'Assign one or more FRPS routes to every Agent from the control plane.', agentTokenHint: 'The enrollment token expires in 10 minutes and can be used only once.',
    nodeCreateHint: 'The permanent credential is written only to the protected Agent configuration on the target OS.', frpsRoutes: 'FRPS routes', addRoute: 'Assign route',
    removeRoute: 'Remove route', unassigned: 'Unassigned', createServerFirst: 'Create FRPS first', selectServer: 'Select FRPS', add: 'Add',
  },
} as const

type Language = keyof typeof messages
type MessageKey = keyof typeof messages.zh
type I18nValue = { language: Language; setLanguage: (language: Language) => void; t: (key: MessageKey) => string }
const I18nContext = createContext<I18nValue | null>(null)

export function I18nProvider({ children }: PropsWithChildren) {
  const [language, setLanguageState] = useState<Language>(() => localStorage.getItem('frp-panel.language') === 'en' ? 'en' : 'zh')
  const value = useMemo<I18nValue>(() => ({
    language,
    setLanguage: (next) => { localStorage.setItem('frp-panel.language', next); setLanguageState(next) },
    t: (key) => messages[language][key],
  }), [language])
  return <I18nContext.Provider value={value}>{children}</I18nContext.Provider>
}

export function useI18n(): I18nValue {
  const value = useContext(I18nContext)
  if (!value) throw new Error('I18nProvider is missing')
  return value
}
