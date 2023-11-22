
<template>
  <Layout v-model="active">
    <template v-if="active === 'profiles'">
      <Profiles ></Profiles>
    </template>
    <template v-else-if="active === 'general'">
      <General />
    </template>
    <template v-else-if="active === 'log'">
      <Log />
    </template>
  </Layout>
</template>

<script setup>
import Layout from './components/layout/index.vue'
import Profiles from './components/profiles/index.vue'
import General from './components/general/index.vue'
import Log from './components/log/index.vue'

import { ref } from 'vue'

let active = ref('profiles')

</script>


<style>
.nlink-ui-content {
  background-color: var(--content-bg-color);
  padding: 10px;
  /* height: 100%; */
}
</style>
