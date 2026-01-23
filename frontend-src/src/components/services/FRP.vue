<template>
  <v-card subtitle="FRP" class="border" rounded="xl">
    <v-card-text>
      <!-- Mode Selection -->
      <v-row>
        <v-col cols="12" sm="6">
          <v-select
            :label="$t('mode')"
            :items="[
              { title: 'Server', value: 'server' },
              { title: 'Client', value: 'client' }
            ]"
            v-model="data.mode"
            hide-details
          ></v-select>
        </v-col>
      </v-row>

      <!-- Server Configuration -->
      <FRPServer v-if="data.mode === 'server'" :data="data" />

      <!-- Client Configuration -->
      <FRPClient v-if="data.mode === 'client'" :data="data" />
    </v-card-text>
  </v-card>
</template>

<script setup lang="ts">
import FRPServer from './FRPServer.vue'
import FRPClient from './FRPClient.vue'

defineProps<{
  data: any
}>()
</script>
