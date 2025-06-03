<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useAuth } from '@/stores/auth'
import axios from 'axios'

const auth = useAuth()

if (!auth.isAuthenticated) {
  window.location.href = '/login'
}

type Server = {
  id: number;
  created_at: string;
  updated_at: string;
  name: string;
  host_name: string;
  organization_id: number;
};

type ServerWithStatus = Server & { status?: string };

const servers = ref<ServerWithStatus[]>([])
const totalCount = ref(0)
const page = ref(1)
const pageSize = ref(10)

const fetchServers = async () => {
  try {
    const response = await axios.get('/api/servers', {
      headers: { 'Authorization': `Bearer ${auth.token}` },
      params: { page: page.value, pageSize: pageSize.value }
    })
    const list: Server[] = response.data.servers
    totalCount.value = response.data.total_count

    const withStatus = await Promise.all(
      list.map(async (s) => {
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
    if (axios.isAxiosError(err) && err.response && err.response.status === 401) {
      console.log('Unauthorized access, logging out...')
      auth.logout()
      return
    }
    console.error('Error fetching servers:', err)
  }
}

const fetchServerById = async (id: number) => {
  const res = await axios.get(`/api/server/${id}`, {
    headers: { 'Authorization': `Bearer ${auth.token}` }
  })
  if (res.status === 401) {
    auth.logout()
    return
  }
  return res.data
}

const postServerPowerOff = (id: number) => {
  return axios.post(`/api/server/${id}/power/off`, {}, {
    headers: { 'Authorization': `Bearer ${auth.token}` }
  })
}

const postServerPowerOn = (id: number) => {
  return axios.post(`/api/server/${id}/power/on`, {}, {
    headers: { 'Authorization': `Bearer ${auth.token}` }
  })
}
const postServerPowerReboot = (id: number) => {
  return axios.post(`/api/server/${id}/power/reboot`, {}, {
    headers: { 'Authorization': `Bearer ${auth.token}` }
  })
}
const postServerPowerForceReboot = (id: number) => {
  return axios.post(`/api/server/${id}/power/force-reboot`, {}, {
    headers: { 'Authorization': `Bearer ${auth.token}` }
  })
}
const postServerPowerForceOff = (id: number) => {
  return axios.post(`/api/server/${id}/power/force-off`, {}, {
    headers: { 'Authorization': `Bearer ${auth.token}` }
  })
}

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

onMounted(() => fetchServers())
</script>

<template>
  <div class="home">
    <h1>home</h1>
  </div>

  <div class="server-list">
    <h2>Server List</h2>
    <ul>
      <li v-for="server in servers" :key="server.id">
        <strong>{{ server.name }}</strong> ({{ server.host_name }}) - Status: {{ server.status }} <button
          @click="postServerPowerOn(server.id)">Power On</button>
        <button @click="postServerPowerOff(server.id)">Power Off</button>
        <button @click="postServerPowerReboot(server.id)">Reboot</button>
        <button @click="postServerPowerForceReboot(server.id)">Force Reboot</button>
        <button @click="postServerPowerForceOff(server.id)">Force Off</button>
      </li>
    </ul>

    <div class="pagination">
      <button @click="previousPage" :disabled="page === 1">Previous</button>
      <span>Page {{ page }}</span>
      <button @click="nextPage" :disabled="page * pageSize >= totalCount">Next</button>
    </div>
  </div>
</template>

<style></style>
