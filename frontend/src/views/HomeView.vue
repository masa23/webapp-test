<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useAuth } from '@/stores/auth'
import axios from 'axios'

const auth = useAuth()

// 型定義
interface Server {
  id: number
  created_at: string
  updated_at: string
  name: string
  host_name: string
  organization_id: number
}

interface ServerWithStatus extends Server {
  status?: string
}

// ステート
const servers = ref<ServerWithStatus[]>([])
const totalCount = ref(0)
const page = ref(1)
const pageSize = ref(10)
const loading = ref(true)

// サーバー一覧取得
const fetchServers = async () => {
  loading.value = true
  try {
    const { data } = await axios.get('/api/servers', {
      headers: { Authorization: `Bearer ${await auth.getToken()}` },
      params: { page: page.value, pageSize: pageSize.value }
    })
    totalCount.value = data.total_count

    const withStatus = await Promise.all(
      data.servers.map(async (s: Server) => {
        try {
          const detail = await fetchServerById(s.id)
          return { ...s, status: detail.status }
        } catch {
          return { ...s, status: 'unknown' }
        }
      })
    )
    servers.value = withStatus
  } catch (err) {
    handleAxiosError(err)
  } finally {
    loading.value = false
  }
}

// サーバー詳細取得
const fetchServerById = async (id: number) => {
  const res = await axios.get(`/api/server/${id}`, {
    headers: { Authorization: `Bearer ${await auth.getToken()}` }
  })
  return res.data
}

// 電源操作共通処理
const postServerAction = async (id: number, action: string, confirmMsg: string, updateStatus = false) => {
  if (!confirm(confirmMsg)) return
  try {
    await axios.post(`/api/server/${id}/${action}`, {}, {
      headers: { Authorization: `Bearer ${await auth.getToken()}` }
    })
    if (updateStatus) {
      setTimeout(async () => {
        try {
          const server = await fetchServerById(id)
          const index = servers.value.findIndex(s => s.id === id)
          if (index !== -1) servers.value[index].status = server.status
        } catch (err) {
          console.error('Error fetching server:', err)
        }
      }, 5000)
    }
  } catch (err) {
    handleAxiosError(err)
  }
}

// 電源操作API
const postServerPowerOff = (id: number) => postServerAction(id, 'power/off', '本当に電源を切りますか？', true)
const postServerPowerOn = (id: number) => postServerAction(id, 'power/on', '本当に電源を入れますか？', true)
const postServerPowerReboot = (id: number) => postServerAction(id, 'power/reboot', '本当に再起動しますか？', true)
const postServerPowerForceReboot = (id: number) => postServerAction(id, 'power/force-reboot', '本当に強制再起動しますか？', true)
const postServerPowerForceOff = (id: number) => postServerAction(id, 'power/force-off', '本当に強制停止しますか？', true)

// エラーハンドリング
const handleAxiosError = (err: unknown) => {
  if (axios.isAxiosError(err) && err.response?.status === 401) {
    console.log('Unauthorized access, logging out...')
    auth.logout()
  } else {
    console.error('API error:', err)
  }
}

const openVNC = async (id: number) => {
  const url = `/noVNC/vnc.html?autoconnect=true&path=/ws/server/${id}/vnc?token=${await auth.getToken()}`
  console.log('Opening VNC:', url)
  // 新しいタブで開く
  window.open(url, '_blank')
}

// ページ操作
const nextPage = () => {
  page.value++
  fetchServers()
}

const previousPage = () => {
  if (page.value > 1) {
    page.value--
    fetchServers()
  }
}

// 初回ロード
onMounted(() => fetchServers())
</script>

<template>
  <div class="p-8 space-y-8">
    <div>
      <h1 class="text-3xl font-bold text-gray-900">Server Dashboard</h1>
      <button @click="fetchServers"
        class="flex items-center gap-1 px-3 py-1.5 bg-blue-600 text-white rounded hover:bg-blue-700 transition">
        🔄 Reload
      </button>
    </div>

    <div class="bg-white rounded-xl shadow-md p-6">
      <h2 class="text-xl font-semibold text-gray-800 mb-4">Server List</h2>

      <div v-if="loading" class="text-center text-gray-500 py-10 text-lg">
        🔄 Loading servers...
      </div>

      <div v-else class="overflow-x-auto">
        <table class="min-w-full text-sm">
          <thead>
            <tr class="bg-gray-50 text-gray-600 uppercase text-xs">
              <th class="px-4 py-3 text-left">Name</th>
              <th class="px-4 py-3 text-left">Host</th>
              <th class="px-4 py-3 text-left">Status</th>
              <th class="px-4 py-3 text-left">Actions</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="server in servers" :key="server.id" class="border-b hover:bg-gray-50">
              <td class="px-4 py-3 font-medium text-gray-900">{{ server.name }}</td>
              <td class="px-4 py-3 text-gray-500">{{ server.host_name }}</td>
              <td class="px-4 py-3">
                <span class="px-2 py-1 rounded-full text-xs font-medium" :class="{
                  'bg-green-100 text-green-700': server.status === 'running',
                  'bg-yellow-100 text-yellow-700': server.status === 'stopped',
                  'bg-gray-200 text-gray-600': server.status === 'unknown',
                }">
                  {{ server.status }}
                </span>
              </td>
              <td class="px-4 py-3">
                <div class="flex flex-wrap gap-2">
                  <button @click="postServerPowerOn(server.id)" class="btn btn-green">On</button>
                  <button @click="postServerPowerOff(server.id)" class="btn btn-yellow">Off</button>
                  <button @click="postServerPowerReboot(server.id)" class="btn btn-blue">Reboot</button>
                  <button @click="postServerPowerForceReboot(server.id)" class="btn btn-red">Force Reboot</button>
                  <button @click="postServerPowerForceOff(server.id)" class="btn btn-dark">Force Off</button>
                  <button @click="openVNC(server.id)" class="btn btn-light">VNC</button>
                </div>
              </td>
            </tr>
          </tbody>
        </table>

        <!-- ページネーション -->
        <div class="flex justify-center items-center mt-6 gap-4">
          <button @click="previousPage" :disabled="page === 1" class="btn btn-light">
            ⬅ Prev
          </button>
          <span class="text-sm text-gray-600">Page {{ page }}</span>
          <button @click="nextPage" :disabled="page * pageSize >= totalCount" class="btn btn-light">
            Next ➡
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<style lang="postcss" scoped>
.btn {
  @apply px-3 py-1 text-sm rounded font-medium transition;
}

.btn-green {
  @apply bg-green-600 text-white hover:bg-green-700;
}

.btn-yellow {
  @apply bg-yellow-500 text-white hover:bg-yellow-600;
}

.btn-blue {
  @apply bg-blue-600 text-white hover:bg-blue-700;
}

.btn-red {
  @apply bg-red-500 text-white hover:bg-red-600;
}

.btn-dark {
  @apply bg-gray-700 text-white hover:bg-gray-800;
}

.btn-light {
  @apply bg-gray-200 text-gray-700 hover:bg-gray-300 disabled:opacity-50;
}
</style>
