<template>
  <v-card rounded="xl" elevation="5" min-width="280" :title="service.tag">
    <template #subtitle>
      <v-row align="center">
        <v-col class="text-caption">
          <v-chip size="x-small" :color="getStatusColor()">
            {{ service.type.toUpperCase() }}
          </v-chip>
          <span v-if="service.mode" class="ml-2">{{ service.mode }}</span>
        </v-col>
      </v-row>
    </template>

    <v-card-text>
      <!-- Server模式显示 -->
      <template v-if="service.mode === 'server'">
        <v-row>
          <v-col cols="5" class="text-caption">{{ $t('in.addr') }}</v-col>
          <v-col cols="7" class="text-right">
            <span class="text-subtitle-2">0.0.0.0:{{ service.bind_port }}</span>
          </v-col>
        </v-row>
        <v-row v-if="service.vhost_http_port">
          <v-col cols="5" class="text-caption">HTTP Port</v-col>
          <v-col cols="7" class="text-right">
            <span class="text-subtitle-2">{{ service.vhost_http_port }}</span>
          </v-col>
        </v-row>
        <v-row v-if="service.dashboard_port">
          <v-col cols="5" class="text-caption">Dashboard</v-col>
          <v-col cols="7" class="text-right">
            <span class="text-subtitle-2">
              {{ service.dashboard_addr || '0.0.0.0' }}:{{ service.dashboard_port }}
            </span>
          </v-col>
        </v-row>
      </template>

      <!-- Client模式显示 -->
      <template v-if="service.mode === 'client'">
        <v-row>
          <v-col cols="5" class="text-caption">{{ $t('out.addr') }}</v-col>
          <v-col cols="7" class="text-right">
            <span class="text-subtitle-2">{{ service.server_addr }}</span>
          </v-col>
        </v-row>
        <v-row>
          <v-col cols="5" class="text-caption">{{ $t('out.port') }}</v-col>
          <v-col cols="7" class="text-right">
            <span class="text-subtitle-2">{{ service.server_port }}</span>
          </v-col>
        </v-row>
        <v-row>
          <v-col cols="5" class="text-caption">Proxies</v-col>
          <v-col cols="7" class="text-right">
            <span class="text-subtitle-2">{{ service.proxies?.length || 0 }}</span>
          </v-col>
        </v-row>
      </template>

      <!-- 运行状态 -->
      <v-row class="mt-2">
        <v-col cols="5" class="text-caption">{{ $t('status') }}</v-col>
        <v-col cols="7" class="text-right">
          <v-chip
            v-if="!loading"
            size="x-small"
            :color="status?.running ? 'success' : 'error'"
          >{{ status?.running ? $t('running') : $t('stopped') }}</v-chip>
          <v-progress-circular
            v-if="loading"
            indeterminate
            size="16"
            color="primary"
          ></v-progress-circular>
        </v-col>
      </v-row>
    </v-card-text>

    <v-divider></v-divider>

    <v-card-actions class="pa-2">
      <!-- 编辑按钮 -->
      <v-btn icon="mdi-file-edit" @click="$emit('edit', service.id)" size="small">
        <v-icon />
        <v-tooltip activator="parent" location="top" :text="$t('actions.edit')"></v-tooltip>
      </v-btn>

      <!-- FRP操作按钮 -->
      <template v-if="status?.running">
        <v-btn icon="mdi-stop-circle" color="error" @click="stopFRP" size="small">
          <v-icon />
          <v-tooltip activator="parent" location="top" :text="$t('stop')"></v-tooltip>
        </v-btn>
        <v-btn icon="mdi-refresh" @click="restartFRP" size="small">
          <v-icon />
          <v-tooltip activator="parent" location="top" :text="$t('restart')"></v-tooltip>
        </v-btn>
      </template>

      <template v-else>
        <v-btn icon="mdi-play-circle" color="success" @click="startFRP" size="small">
          <v-icon />
          <v-tooltip activator="parent" location="top" :text="$t('start')"></v-tooltip>
        </v-btn>
      </template>

      <!-- 日志按钮 -->
      <v-btn icon="mdi-file-document-outline" @click="showLogs = true" size="small">
        <v-icon />
        <v-tooltip activator="parent" location="top" :text="$t('logs')"></v-tooltip>
      </v-btn>

      <!-- 删除按钮 -->
      <v-btn
        icon="mdi-file-remove"
        color="warning"
        @click="$emit('delete', service.id)"
        style="margin-inline-start: auto;"
        size="small"
      >
        <v-icon />
        <v-tooltip activator="parent" location="top" :text="$t('actions.del')"></v-tooltip>
      </v-btn>
    </v-card-actions>

    <!-- 日志对话框 -->
    <v-dialog v-model="showLogs" max-width="800">
      <v-card>
        <v-card-title>
          {{ $t('logs') }} - {{ service.tag }}
          <v-spacer></v-spacer>
          <v-btn icon="mdi-refresh" @click="loadLogs" :loading="logsLoading"></v-btn>
          <v-btn icon="mdi-close" @click="showLogs = false"></v-btn>
        </v-card-title>
        <v-divider></v-divider>
        <v-card-text style="max-height: 500px; overflow-y: auto;">
          <pre v-if="logs.length > 0" class="text-caption">{{ logs.join('\n') }}</pre>
          <v-alert v-else type="info">{{ $t('noLogs') }}</v-alert>
        </v-card-text>
      </v-card>
    </v-dialog>
  </v-card>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'
import { FRP } from '@/types/services'
import Data from '@/store/modules/data'

const props = defineProps<{
  service: FRP
}>()

const emit = defineEmits<{
  edit: [id: number]
  delete: [id: number]
}>()

const loading = ref(false)
const showLogs = ref(false)
const logsLoading = ref(false)
const logs = ref<string[]>([])

const status = ref<any>(null)
let statusInterval: ReturnType<typeof setInterval> | null = null

// 加载状态
const loadStatus = async () => {
  loading.value = true
  try {
    const baseURL = (window as any).BASE_URL || '/'
    const resp = await fetch(`${baseURL}api/frp/status?id=${props.service.id}`, {
      credentials: 'include'
    })
    const data = await resp.json()
    if (data.success) {
      status.value = data.obj
    }
  } catch (error) {
    console.error('Failed to load FRP status:', error)
  } finally {
    loading.value = false
  }
}

// 加载日志
const loadLogs = async () => {
  logsLoading.value = true
  try {
    const baseURL = (window as any).BASE_URL || '/'
    const resp = await fetch(`${baseURL}api/frp/logs?id=${props.service.id}&lines=100`, {
      credentials: 'include'
    })
    const data = await resp.json()
    if (data.success) {
      logs.value = data.obj || []
    }
  } catch (error) {
    console.error('Failed to load FRP logs:', error)
  } finally {
    logsLoading.value = false
  }
}

const getStatusColor = () => {
  if (!status.value) return 'grey'
  return status.value.running ? 'success' : 'error'
}

// FRP 操作
const startFRP = async () => {
  loading.value = true
  try {
    const formData = new FormData()
    formData.append('id', String(props.service.id))

    const baseURL = (window as any).BASE_URL || '/'
    const resp = await fetch(`${baseURL}api/frp/start`, {
      method: 'POST',
      credentials: 'include',
      body: formData
    })
    const data = await resp.json()
    if (data.success) {
      await loadStatus()
    } else {
      alert('Failed to start FRP: ' + data.msg)
    }
  } catch (error) {
    console.error('Failed to start FRP:', error)
    alert('Failed to start FRP: ' + error)
  } finally {
    loading.value = false
  }
}

const stopFRP = async () => {
  loading.value = true
  try {
    const formData = new FormData()
    formData.append('id', String(props.service.id))

    const baseURL = (window as any).BASE_URL || '/'
    const resp = await fetch(`${baseURL}api/frp/stop`, {
      method: 'POST',
      credentials: 'include',
      body: formData
    })
    const data = await resp.json()
    if (data.success) {
      await loadStatus()
    } else {
      alert('Failed to stop FRP: ' + data.msg)
    }
  } catch (error) {
    console.error('Failed to stop FRP:', error)
    alert('Failed to stop FRP: ' + error)
  } finally {
    loading.value = false
  }
}

const restartFRP = async () => {
  loading.value = true
  try {
    const formData = new FormData()
    formData.append('id', String(props.service.id))

    const baseURL = (window as any).BASE_URL || '/'
    const resp = await fetch(`${baseURL}api/frp/restart`, {
      method: 'POST',
      credentials: 'include',
      body: formData
    })
    const data = await resp.json()
    if (data.success) {
      await loadStatus()
    } else {
      alert('Failed to restart FRP: ' + data.msg)
    }
  } catch (error) {
    console.error('Failed to restart FRP:', error)
    alert('Failed to restart FRP: ' + error)
  } finally {
    loading.value = false
  }
}

// 组件挂载时加载状态并设置定时刷新
onMounted(() => {
  loadStatus()
  statusInterval = setInterval(loadStatus, 5000)
})

// 组件卸载时清除定时器
onUnmounted(() => {
  if (statusInterval) {
    clearInterval(statusInterval)
    statusInterval = null
  }
})
</script>
