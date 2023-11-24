<template>
  <div class="nlink-ui-content nlink-ui-profiles">
    <div class="nlink-ui-link-import">
      <!-- 链接引入 -->
      <div class="nlink-ui-link-input">
        <n-input
          v-model:value="link" round placeholder="填入订阅链接">
          <template #suffix>
            <n-icon :component="FlashOutline" />
          </template>
        </n-input>
      </div>
      <div class="nlink-ui-link-operate">
        <button @click="tryImport">导入</button>
        <button @click="updateAll">更新全部</button>
      </div>
    </div>
    <div class="nlink-ui-subscriptions">
      <div
        class="nlink-ui-subscription"
        :class="{ 'active': profiler.currentProfile.id === item.id }"
        v-for="item in profiler.profiles" :key="item.id"
        @click="setCurrentProfile(item)">
        <div>
          <div class="">{{ item.name }}</div>
          <div>{{ item.type }}({{ item.lastUpdatedAt }})</div>
        </div>
      </div>
    </div>
    <div class="nlink-ui-setting">
      <div class="nlink-ui-setting-operate">
        <button @click="changeCode">提交</button>
      </div>
      <div id="nlink-ui-code">
        <!-- :indent-with-tab="true" 是否自动获取焦点-->
        <codemirror v-model="code"
          placeholder="Code gose here..." :style="{ height: '400px' }"
          :autofocus="true"
          :tabSize="2" :extensions="extensions" />
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import { useProfilerStore } from '../../store/index.js'
import { getDefaultProfile, requestRemoteProfile } from '../../utils/profile'
import { ipcEmit } from '../../../ipc'

import { Codemirror } from "vue-codemirror"
import { javascript } from "@codemirror/lang-javascript"
import { oneDark } from "@codemirror/theme-one-dark"
import { EditorView } from "@codemirror/view"
import { FlashOutline } from '@vicons/ionicons5'
import { NInput } from 'naive-ui'

const profiler = useProfilerStore()

let link = ref(null)

// 导入
const tryImport = async () => {
  console.log('tryImport')
  if (!link)  window.$message.warning('请填写链接')

  const profile = requestRemoteProfile(link)
  profiler.pushProfile(profile)
}

if (profiler.profiles.length === 0) {
  profiler.pushProfile(getDefaultProfile())
}

const updateAll = () => {
  console.log('updateAll')
}

// 设置当前配置
const setCurrentProfile = (item) => {
  profiler.setCurrentProfile(item)
  code.value = item.content
}

// 编写配置
let code = ref(profiler.currentProfile.content)
let myTheme = EditorView.theme({
    // 输入的字体颜色
    "&": {
        color: "#0052D9",
        backgroundColor: "#FFFFFF"
    },
    ".cm-content": {
        caretColor: "#0052D9",
    },
    // 激活背景色
    ".cm-activeLine": {
        backgroundColor: "#FAFAFA"
    },
    // 激活序列的背景色
    ".cm-activeLineGutter": {
        backgroundColor: "#FAFAFA"
    },
    //光标的颜色
    "&.cm-focused .cm-cursor": {
        borderLeftColor: "#0052D9"
    },
    // 选中的状态
    "&.cm-focused .cm-selectionBackground, ::selection": {
        backgroundColor: "#0052D9",
        color:'#FFFFFF'
    },
    // 左侧侧边栏的颜色
    ".cm-gutters": {
        backgroundColor: "#FFFFFF",
        color: "#ddd", //侧边栏文字颜色
        border: "none"
    }
}, { dark: true })
const extensions = [javascript(), myTheme];

const changeCode = () => {
  profiler.updateCurrentProfile(code.value)
}

</script>

<style scoped>

.nlink-ui-link-import {
  display: flex;
  padding-bottom: 25px;
  border-bottom: solid 2px var(--bg-color);
}
.nlink-ui-link-input {
  flex: 1;
}
.nlink-ui-link-operate {
  width: 150px;
}
.nlink-ui-link-input input {
  width: 90%;
  padding: 6px 10px;
  border: solid 1px var(--block-bg-color);
  border-radius: 4px;
  background-color: #fff;
  font-size: 20px;
  color: #999;
  outline-style: none;
}
.nlink-ui-link-input input:focus {
  border-color: var(--active-color);
}

.nlink-ui-subscriptions {
  display: flex;
  margin-top: 20px
}

.nlink-ui-subscription {
  width: 40%;
  border-left: solid 2px transparent;
  border-radius: 4px;
  background-color: var(--block-bg-color);
  padding: 10px;
  margin-left: 10px;
  position: relative;
  cursor: pointer;
}
.nlink-ui-subscription.active::before {
  position: absolute;
  content: ' ';
  background-color: var(--active-color);
  height: 100%;
  width: 4px;
  left: -10px;
  top: 0
}

.nlink-ui-setting-operate {
  text-align: right;
  padding: 10px;
}
</style>