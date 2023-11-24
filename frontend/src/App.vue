
<template>
  <n-config-provider :locale="zhCN" :date-locale="dateZhCN" style="height: 100%;">

    <n-message-provider style="height: 100%;">
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
        <template v-else-if="active === 'proxies'">
          <Proxies />
        </template>
      </Layout>
    </n-message-provider>
  </n-config-provider>
</template>

<script setup>
import Layout from './components/layout/index.vue'
import Profiles from './components/profiles/index.vue'
import General from './components/general/index.vue'
import Log from './components/log/index.vue'
import Proxies from './components/proxies/index.vue'

import { NConfigProvider, useMessage } from 'naive-ui'
import { zhCN, dateZhCN } from 'naive-ui'
// import hljs from 'highlight.js/lib/core'
// import shell from 'highlight.js/lib/languages/plaintext'

import { ref, onMounted } from 'vue'

let active = ref('general')

// hljs.registerLanguage('shell', shell);

</script>


<style>
.nlink-ui-content {
  background-color: var(--content-bg-color);
  padding: 10px;
  /* height: 100%; */
}
</style>
