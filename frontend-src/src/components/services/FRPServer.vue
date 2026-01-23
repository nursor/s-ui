<template>
  <div>
    <!-- Basic Server Configuration -->
    <v-card-title class="text-subtitle-1">Basic Configuration</v-card-title>
    <v-row>
      <v-col cols="12" sm="6">
        <v-text-field
          label="Bind Port"
          type="number"
          v-model.number="data.bind_port"
          hide-details
          placeholder="7000"
        ></v-text-field>
      </v-col>
      <v-col cols="12" sm="6">
        <v-text-field
          label="Token (optional)"
          v-model="data.token"
          hide-details
          type="password"
        ></v-text-field>
      </v-col>
    </v-row>

    <!-- Virtual Host Ports -->
    <v-card-title class="text-subtitle-1 mt-2">Virtual Host</v-card-title>
    <v-row>
      <v-col cols="12" sm="6">
        <v-text-field
          label="VHost HTTP Port (optional)"
          type="number"
          v-model.number="data.vhost_http_port"
          hide-details
          placeholder="8080"
        ></v-text-field>
      </v-col>
      <v-col cols="12" sm="6">
        <v-text-field
          label="VHost HTTPS Port (optional)"
          type="number"
          v-model.number="data.vhost_https_port"
          hide-details
          placeholder="8443"
        ></v-text-field>
      </v-col>
    </v-row>

    <!-- Dashboard Configuration -->
    <v-card-title class="text-subtitle-1 mt-2">
      <v-row class="align-center">
        <v-col>Dashboard</v-col>
        <v-col cols="auto">
          <v-switch
            v-model="dashboardEnabled"
            color="primary"
            hide-details
            density="compact"
          ></v-switch>
        </v-col>
      </v-row>
    </v-card-title>
    <v-row v-if="dashboardEnabled">
      <v-col cols="12" sm="6">
        <v-text-field
          label="Dashboard Address"
          v-model="data.dashboard_addr"
          hide-details
          placeholder="0.0.0.0"
        ></v-text-field>
      </v-col>
      <v-col cols="12" sm="6">
        <v-text-field
          label="Dashboard Port"
          type="number"
          v-model.number="data.dashboard_port"
          hide-details
          placeholder="7500"
        ></v-text-field>
      </v-col>
      <v-col cols="12" sm="6">
        <v-text-field
          label="Dashboard User"
          v-model="data.dashboard_user"
          hide-details
        ></v-text-field>
      </v-col>
      <v-col cols="12" sm="6">
        <v-text-field
          label="Dashboard Password"
          v-model="data.dashboard_pwd"
          hide-details
          type="password"
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
          label="Max Pool Count (optional)"
          type="number"
          v-model.number="data.max_pool_count"
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
        <v-text-field
          label="User Conn Timeout (optional, seconds)"
          type="number"
          v-model.number="data.user_conn_timeout"
          hide-details
        ></v-text-field>
      </v-col>
    </v-row>

    <!-- Prometheus Metrics -->
    <v-card-title class="text-subtitle-1 mt-2">
      <v-row class="align-center">
        <v-col>Prometheus Metrics</v-col>
        <v-col cols="auto">
          <v-switch
            v-model="data.enable_prometheus"
            color="primary"
            hide-details
            density="compact"
          ></v-switch>
        </v-col>
      </v-row>
    </v-card-title>
    <v-row v-if="data.enable_prometheus">
      <v-col cols="12" sm="6">
        <v-text-field
          label="Metrics Address"
          v-model="data.metrics_addr"
          hide-details
          placeholder="0.0.0.0"
        ></v-text-field>
      </v-col>
      <v-col cols="12" sm="6">
        <v-text-field
          label="Metrics Port"
          type="number"
          v-model.number="data.metrics_port"
          hide-details
        ></v-text-field>
      </v-col>
    </v-row>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'

const props = defineProps<{
  data: any
}>()

const dashboardEnabled = computed({
  get() {
    return props.data.dashboard_port !== undefined
  },
  set(v: boolean) {
    if (v) {
      props.data.dashboard_port = 7500
      props.data.dashboard_user = 'admin'
      props.data.dashboard_pwd = ''
      props.data.dashboard_addr = '0.0.0.0'
    } else {
      delete props.data.dashboard_port
      delete props.data.dashboard_user
      delete props.data.dashboard_pwd
      delete props.data.dashboard_addr
    }
  }
})
</script>
