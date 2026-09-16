import { createContext, type PropsWithChildren, useContext, useMemo, useState } from 'react'

const zh = {
  overview: '概览', clients: 'Clients', client: 'Client', servers: 'Servers', server: 'Server', tunnels: 'Tunnels', account: '账户设置',
  signOut: '退出登录', language: '界面语言', chinese: '简体中文', english: 'English', online: '在线', offline: '离线',
  createClient: '添加 Client', createServer: '创建 Server', createTunnel: '创建 Tunnel', refresh: '刷新', cancel: '取消', confirm: '确认', close: '关闭', saving: '正在保存…',
  clientId: 'Client ID', serverId: 'Server ID', serverIp: 'Server 公网地址', installAgent: '安装 Client Agent', copy: '复制安装命令', copied: '已复制',
  linux: 'Linux', macos: 'macOS', windows: 'Windows', operatingSystem: '目标操作系统', noData: '暂无数据', loading: '加载中…',
  login: '登录', username: '用户名', usernameOrEmail: '用户名或邮箱', password: '密码', email: '邮箱',
  bootstrapOwner: '首次部署？创建所有者', createOwner: '创建所有者并进入控制台', backToLogin: '返回登录', confirmPassword: '确认密码', passwordMismatch: '两次输入的新密码不一致',
  initialSetup: '首次初始化', accountLogin: '账户登录', setupTitle: '创建首个所有者账户', setupHint: '该账户拥有最高权限。创建成功后会自动登录，浏览器也可正确识别并保存凭据。', loginHint: '登录 Master，统一管理 Server、Client 与 Tunnel。', setupDisabled: '系统尚未创建所有者，但账户初始化已关闭。请暂时启用 APP_ENABLE_REGISTER 后重新加载。',
  passwordChanged: '密码已更新，请使用新密码重新登录。', welcome: '安全地管理你的 FRP 网络', controlPlane: 'Master 控制面', masterConsole: 'Master 控制台',
  capabilityHint: 'Master 只保存期望状态；代理流量由 Client 与 Server 直接传输。', status: '状态', lifecycle: '生命周期', ephemeral: '临时', managed: '受管', lastSeen: '最近上线', agentStatus: 'Agent 状态', bindPort: 'FRPS 绑定端口',
  infrastructure: '基础设施', dataPlane: '数据转发', security: '安全与账户', nextStep: '下一步',
  deployServer: '部署 Server（FRPS）', createDeployment: '创建并生成部署文件', copyCompose: '复制 compose.yaml',
  serverTopologyHint: '每台公网服务器运行一个独立的 Server（FRPS）；Master 不承载代理流量。',
  serverTokenHint: '将 compose.yaml 保存到目标服务器。一次性令牌 10 分钟有效，长期凭据只保存在 Server 数据卷中。',
  serverCreateHint: 'Client 会连接此公网地址；请放行 FRPS 绑定端口及各 Tunnel 使用的公网端口。',
  clientTopologyHint: 'Client 是运行 Client Agent 与受管 FRPC 的业务主机；连接哪个 Server 由每条 Tunnel 决定。',
  agentTokenHint: '选择目标系统并执行命令。注册令牌 10 分钟有效且只能使用一次。',
  clientCreateHint: '这里只注册物理 Client。创建 Tunnel 时再选择目标 Server，系统会自动建立并复用 FRPC 连接。',
  installCommandHint: '命令会把程序、配置和服务写入系统规范目录，不会污染当前目录。',
  clientInventory: 'Client 清单', clientInventoryHint: '列表仅展示物理 Client，不展示系统自动维护的内部 FRPC 连接。', noClients: '尚未添加 Client', noClientsHint: '添加 Client 后，在目标主机执行一次性安装命令。',
  tunnelName: 'Tunnel 名称', protocol: '协议', localService: '本地服务', publicEndpoint: '公网入口', actions: '操作', delete: '删除',
  localHost: '本地地址', localPort: '本地端口', remotePort: '公网端口', noTunnels: '尚未创建 Tunnel', noTunnelsHint: '每条 Tunnel 可以独立选择 Client 与 Server。',
  tunnelTopologyHint: '每条 Tunnel 直接绑定一个 Client 与一个 Server；系统自动管理底层 FRPC 连接。',
  tunnelPrerequisites: '创建 Tunnel 前还缺少资源', createClientFirst: '请先添加至少一个 Client。', createServerFirst: '请先创建至少一个 Server。',
  tunnelCreateHint: '公网端口监听在所选 Server 上，并转发到所选 Client 能访问的本地服务。', deleteTunnelConfirm: '确定删除这条 Tunnel 吗？',
  tunnelInventory: 'Tunnel 清单', tunnelInventoryHint: '一台 Client 可创建多条 Tunnel，每条可使用不同 Server。',
  accountHint: '查看当前账户并安全地更新登录密码。', profile: '账户信息', profileHint: '当前登录身份与权限。', role: '角色',
  changePassword: '修改密码', changePasswordHint: '修改后当前会话将退出，需要使用新密码重新登录。', currentPassword: '当前密码', newPassword: '新密码', passwordPolicy: '新密码至少 12 个字符，且不能与当前密码相同。', updatePassword: '更新密码',
  managedClients: '受管 Client', managedIngress: '受管 Server', desiredTunnels: '期望 Tunnel', secureDefaults: '安全默认配置已启用', secureDefaultsHint: '远程 Shell、Functions 与 WireGuard 等高权限能力默认关闭，需要显式启用。',
  serverInventory: 'Server 清单', serverInventoryHint: '每个 Server 对应一套独立部署的 FRPS。', noServers: '尚未创建 Server', noServersHint: '创建 Server 后，将生成可直接保存的 compose.yaml。',
} as const

type MessageKey = keyof typeof zh
const en: Record<MessageKey, string> = {
  overview: 'Overview', clients: 'Clients', client: 'Client', servers: 'Servers', server: 'Server', tunnels: 'Tunnels', account: 'Account settings',
  signOut: 'Sign out', language: 'Interface language', chinese: '简体中文', english: 'English', online: 'Online', offline: 'Offline',
  createClient: 'Add Client', createServer: 'Create Server', createTunnel: 'Create Tunnel', refresh: 'Refresh', cancel: 'Cancel', confirm: 'Confirm', close: 'Close', saving: 'Saving…',
  clientId: 'Client ID', serverId: 'Server ID', serverIp: 'Server public address', installAgent: 'Install Client Agent', copy: 'Copy install command', copied: 'Copied',
  linux: 'Linux', macos: 'macOS', windows: 'Windows', operatingSystem: 'Target operating system', noData: 'No data yet', loading: 'Loading…',
  login: 'Sign in', username: 'Username', usernameOrEmail: 'Username or email', password: 'Password', email: 'Email',
  bootstrapOwner: 'First run? Create owner', createOwner: 'Create owner and open console', backToLogin: 'Back to sign in', confirmPassword: 'Confirm password', passwordMismatch: 'The new passwords do not match',
  initialSetup: 'Initial setup', accountLogin: 'Account login', setupTitle: 'Create the first owner account', setupHint: 'This account has full access. You will be signed in after creation, and password managers can save the credentials correctly.', loginHint: 'Sign in to Master to manage Servers, Clients, and Tunnels.', setupDisabled: 'No owner exists, but account setup is disabled. Temporarily enable APP_ENABLE_REGISTER and reload this page.',
  passwordChanged: 'Password updated. Sign in again with the new password.', welcome: 'Manage your FRP network securely', controlPlane: 'Master control plane', masterConsole: 'Master console',
  capabilityHint: 'Master stores desired state only; proxy traffic flows directly between Clients and Servers.', status: 'Status', lifecycle: 'Lifecycle', ephemeral: 'Ephemeral', managed: 'Managed', lastSeen: 'Last seen', agentStatus: 'Agent status', bindPort: 'FRPS bind port',
  infrastructure: 'Infrastructure', dataPlane: 'Data forwarding', security: 'Security and account', nextStep: 'Next step',
  deployServer: 'Deploy Server (FRPS)', createDeployment: 'Create deployment', copyCompose: 'Copy compose.yaml',
  serverTopologyHint: 'Each public host runs an independent Server (FRPS); Master carries no proxy traffic.',
  serverTokenHint: 'Save compose.yaml on the target host. The one-use token expires in 10 minutes; long-lived credentials remain only in the Server data volume.',
  serverCreateHint: 'Clients connect to this public address. Allow the FRPS bind port and every Tunnel public port.',
  clientTopologyHint: 'A Client is a workload host running the Client Agent and managed FRPC; each Tunnel chooses its Server.',
  agentTokenHint: 'Select the target OS and run the command. The enrollment token expires in 10 minutes and can be used once.',
  clientCreateHint: 'This registers a physical Client only. Select a Server when creating a Tunnel; the system creates and reuses FRPC connections automatically.',
  installCommandHint: 'The installer uses OS-owned program, configuration, and service locations and leaves the current directory clean.',
  clientInventory: 'Client inventory', clientInventoryHint: 'Only physical Clients are shown; internal FRPC connections are maintained automatically.', noClients: 'No Clients yet', noClientsHint: 'Add a Client, then run its one-use command on the target host.',
  tunnelName: 'Tunnel name', protocol: 'Protocol', localService: 'Local service', publicEndpoint: 'Public endpoint', actions: 'Actions', delete: 'Delete',
  localHost: 'Local host', localPort: 'Local port', remotePort: 'Public port', noTunnels: 'No Tunnels yet', noTunnelsHint: 'Every Tunnel can independently select a Client and Server.',
  tunnelTopologyHint: 'Every Tunnel binds one Client directly to one Server; the system manages the underlying FRPC connection.',
  tunnelPrerequisites: 'Resources required before creating a Tunnel', createClientFirst: 'Add at least one Client first.', createServerFirst: 'Create at least one Server first.',
  tunnelCreateHint: 'The public port listens on the selected Server and forwards to a local service reachable by the selected Client.', deleteTunnelConfirm: 'Delete this Tunnel?',
  tunnelInventory: 'Tunnel inventory', tunnelInventoryHint: 'A Client can own multiple Tunnels, and each can use a different Server.',
  accountHint: 'Review the signed-in account and update its login password securely.', profile: 'Account information', profileHint: 'Current identity and access role.', role: 'Role',
  changePassword: 'Change password', changePasswordHint: 'This signs out the current session; sign in again with the new password.', currentPassword: 'Current password', newPassword: 'New password', passwordPolicy: 'Use at least 12 characters and choose a password different from the current one.', updatePassword: 'Update password',
  managedClients: 'Managed Clients', managedIngress: 'Managed Servers', desiredTunnels: 'Desired Tunnels', secureDefaults: 'Secure defaults are active', secureDefaultsHint: 'Privileged capabilities such as remote Shell, Functions, and WireGuard are disabled until explicitly enabled.',
  serverInventory: 'Server inventory', serverInventoryHint: 'Each Server represents an independently deployed FRPS.', noServers: 'No Servers yet', noServersHint: 'Create a Server to generate a ready-to-save compose.yaml.',
}

const messages = { zh, en }
type Language = keyof typeof messages
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
