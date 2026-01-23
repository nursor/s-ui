<template>
  <div>
    <v-row>
      <v-col cols="12" sm="6">
        <v-text-field
          label="Proxy Name"
          v-model="data.name"
          hide-details
          placeholder="ssh"
        ></v-text-field>
      </v-col>
      <v-col cols="12" sm="6">
        <v-select
          label="Proxy Type"
          :items="[
            { title: 'TCP', value: 'tcp' },
            { title: 'UDP', value: 'udp' },
            { title: 'HTTP', value: 'http' },
            { title: 'HTTPS', value: 'https' },
            { title: 'STCP', value: 'stcp' },
            { title: 'XTCP', value: 'xtcp' }
          ]"
          v-model="data.type"
          hide-details
        ></v-select>
      </v-col>
    </v-row>

    <!-- Local Configuration -->
    <v-row class="mt-2">
      <v-col cols="12" sm="6">
        <v-text-field
          label="Local IP"
          v-model="data.local_ip"
          hide-details
          placeholder="127.0.0.1"
        ></v-text-field>
      </v-col>
      <v-col cols="12" sm="6">
        <v-text-field
          label="Local Port"
          type="number"
          v-model.number="data.local_port"
          hide-details
        ></v-text-field>
      </v-col>
    </v-row>

    <!-- Remote Port (for TCP/UDP) -->
    <v-row v-if="data.type === 'tcp' || data.type === 'udp'">
      <v-col cols="12" sm="6">
        <v-text-field
          label="Remote Port"
          type="number"
          v-model.number="data.remote_port"
          hide-details
        ></v-text-field>
      </v-col>
    </v-row>

    <!-- HTTP/HTTPS Configuration -->
    <v-row v-if="data.type === 'http' || data.type === 'https'">
      <v-col cols="12" sm="6">
        <v-text-field
          label="Custom Domains"
          v-model="data.custom_domains"
          hide-details
          placeholder="example.com,www.example.com"
        ></v-text-field>
      </v-col>
      <v-col cols="12" sm="6">
        <v-text-field
          label="Subdomain (optional)"
          v-model="data.subdomain"
          hide-details
        ></v-text-field>
      </v-col>
      <v-col cols="12" sm="6">
        <v-text-field
          label="Locations (optional)"
          v-model="data.locations"
          hide-details
        ></v-text-field>
      </v-col>
      <v-col cols="12" sm="6">
        <v-text-field
          label="Host Header Rewrite (optional)"
          v-model="data.host_header_rewrite"
          hide-details
        ></v-text-field>
      </v-col>
    </v-row>

    <!-- STCP/XTCP Configuration -->
    <v-row v-if="data.type === 'stcp' || data.type === 'xtcp'">
      <v-col cols="12" sm="6">
        <v-text-field
          label="Secret Key"
          v-model="data.sk"
          hide-details
        ></v-text-field>
      </v-col>
    </v-row>

    <!-- Advanced Options -->
    <v-row class="mt-2">
      <v-col cols="12" sm="6">
        <v-switch
          v-model="data.use_encryption"
          label="Use Encryption"
          color="primary"
          hide-details
          density="compact"
        ></v-switch>
      </v-col>
      <v-col cols="12" sm="6">
        <v-switch
          v-model="data.use_compression"
          label="Use Compression"
          color="primary"
          hide-details
          density="compact"
        ></v-switch>
      </v-col>
    </v-row>

    <!-- Bandwidth Limit (optional) -->
    <v-row>
      <v-col cols="12" sm="6">
        <v-text-field
          label="Bandwidth Limit (optional, e.g., 1MB)"
          v-model="data.bandwidth_limit"
          hide-details
        ></v-text-field>
      </v-col>
    </v-row>

    <!-- Health Check (HTTP only) -->
    <v-row v-if="data.type === 'http' || data.type === 'https'">
      <v-col cols="12">
        <v-divider class="my-2"></v-divider>
        <div class="text-caption text-grey">Health Check (Optional)</div>
      </v-col>
      <v-col cols="12" sm="6">
        <v-select
          label="Health Check Type"
          :items="['tcp', 'http']"
          v-model="data.health_check_type"
          hide-details
        ></v-select>
      </v-col>
      <v-col cols="12" sm="6" v-if="data.health_check_type === 'http'">
        <v-text-field
          label="Health Check URL"
          v-model="data.health_check_url"
          hide-details
          placeholder="/status"
        ></v-text-field>
      </v-col>
      <v-col cols="12" sm="6">
        <v-text-field
          label="Check Interval (seconds)"
          type="number"
          v-model.number="data.health_check_interval"
          hide-details
        ></v-text-field>
      </v-col>
      <v-col cols="12" sm="6">
        <v-text-field
          label="Max Failed"
          type="number"
          v-model.number="data.health_check_max_failed"
          hide-details
        ></v-text-field>
      </v-col>
    </v-row>
  </div>
</template>

<script setup lang="ts">
defineProps<{
  data: any
}>()
</script>
