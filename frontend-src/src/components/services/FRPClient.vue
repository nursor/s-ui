<template>
  <div>
    <!-- Server Connection -->
    <v-card-title class="text-subtitle-1">Server Connection</v-card-title>
    <v-row>
      <v-col cols="12" sm="8">
        <v-text-field
          label="Server Address"
          v-model="data.server_addr"
          hide-details
          placeholder="127.0.0.1"
        ></v-text-field>
      </v-col>
      <v-col cols="12" sm="4">
        <v-text-field
          label="Server Port"
          type="number"
          v-model.number="data.server_port"
          hide-details
          placeholder="7000"
        ></v-text-field>
      </v-col>
    </v-row>
    <v-row>
      <v-col cols="12" sm="6">
        <v-text-field
          label="Token (optional)"
          v-model="data.token"
          hide-details
          type="password"
        ></v-text-field>
      </v-col>
      <v-col cols="12" sm="6">
        <v-text-field
          label="User (optional)"
          v-model="data.user"
          hide-details
        ></v-text-field>
      </v-col>
    </v-row>

    <!-- Advanced Options -->
    <v-card-title class="text-subtitle-1 mt-2">Advanced Options</v-card-title>
    <v-row>
      <v-col cols="12" sm="6">
        <v-switch
          v-model="data.tcp_mux"
          label="TCP Multiplexing"
          color="primary"
          hide-details
        ></v-switch>
      </v-col>
      <v-col cols="12" sm="6">
        <v-text-field
          label="Pool Count (optional)"
          type="number"
          v-model.number="data.pool_count"
          hide-details
          placeholder="1"
        ></v-text-field>
      </v-col>
      <v-col cols="12" sm="6">
        <v-text-field
          label="Heartbeat Interval (optional, seconds)"
          type="number"
          v-model.number="data.heartbeat_interval"
          hide-details
        ></v-text-field>
      </v-col>
      <v-col cols="12" sm="6">
        <v-text-field
          label="Heartbeat Timeout (optional, seconds)"
          type="number"
          v-model.number="data.heartbeat_timeout"
          hide-details
        ></v-text-field>
      </v-col>
      <v-col cols="12" sm="6">
        <v-switch
          v-model="data.login_fail_exit"
          label="Exit on Login Failure"
          color="primary"
          hide-details
        ></v-switch>
      </v-col>
      <v-col cols="12" sm="6">
        <v-text-field
          label="DNS Server (optional)"
          v-model="data.dns_server"
          hide-details
        ></v-text-field>
      </v-col>
    </v-row>

    <!-- Proxies Configuration -->
    <v-card-title class="text-subtitle-1 mt-2">
      <v-row class="align-center">
        <v-col>Proxies</v-col>
        <v-col cols="auto">
          <v-chip
            color="primary"
            density="compact"
            variant="elevated"
            @click="addProxy"
          >
            <v-icon icon="mdi-plus" />
          </v-chip>
        </v-col>
      </v-row>
    </v-card-title>

    <v-card
      v-for="(proxy, index) in data.proxies"
      :key="index"
      class="border mb-2"
      rounded="xl"
    >
      <v-card-text>
        <v-row>
          <v-col cols="auto" align-self="center">
            <v-icon
              @click="removeProxy(index)"
              color="error"
              icon="mdi-delete"
            ></v-icon>
          </v-col>
          <v-col>
            <FRPProxy :data="proxy" />
          </v-col>
        </v-row>
      </v-card-text>
    </v-card>

    <v-alert v-if="!data.proxies || data.proxies.length === 0" type="info" variant="tonal" class="mt-2">
      No proxies configured. Click the + button to add a proxy.
    </v-alert>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import FRPProxy from './FRPProxy.vue'

const props = defineProps<{
  data: any
}>()

const addProxy = () => {
  if (!props.data.proxies) {
    props.data.proxies = []
  }
  props.data.proxies.push({
    name: `proxy${props.data.proxies.length + 1}`,
    type: 'tcp',
    local_ip: '127.0.0.1',
    local_port: 0,
    remote_port: 0
  })
}

const removeProxy = (index: number) => {
  props.data.proxies.splice(index, 1)
}
</script>
