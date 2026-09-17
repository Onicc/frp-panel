import { createRouter, createWebHistory } from 'vue-router'
import { useAuthStore } from '../stores/auth'
import LoginView from '../views/LoginView.vue'
import OverviewView from '../views/OverviewView.vue'
import ClientsView from '../views/ClientsView.vue'
import ServersView from '../views/ServersView.vue'
import TunnelsView from '../views/TunnelsView.vue'
import AccountView from '../views/AccountView.vue'

const router=createRouter({history:createWebHistory(),routes:[
  {path:'/login',component:LoginView,meta:{public:true,title:'auth.signIn'}},
  {path:'/',component:OverviewView,meta:{title:'nav.overview'}},
  {path:'/clients',component:ClientsView,meta:{title:'nav.clients'}},
  {path:'/servers',component:ServersView,meta:{title:'nav.servers'}},
  {path:'/tunnels',component:TunnelsView,meta:{title:'nav.tunnels'}},
  {path:'/account',component:AccountView,meta:{title:'account.title'}},
  {path:'/:pathMatch(.*)*',redirect:'/'}
]})
router.beforeEach(async(to)=>{const auth=useAuthStore();if(!auth.initialized)await auth.load();if(to.path==='/login'&&auth.user)return '/';if(to.meta.public)return true;if(!auth.user)return {path:'/login',query:{redirect:to.fullPath}};return true})
export default router
