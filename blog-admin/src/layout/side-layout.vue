<script setup lang="ts">
import { useAppConfigStore } from '@/stores/app-config'
import { useLocalUserStore } from '@/stores/user'
import { useRoute, useRouter } from 'vue-router'
import { useArcoMessage } from '@/hooks/message'
import { onMounted, ref, watch } from 'vue'
import * as ArcoIcons from '@arco-design/web-vue/es/icon';

const router = useRouter()
const route = useRoute();
const { errorMessage } = useArcoMessage();
const configStore = useAppConfigStore();
const userStore = useLocalUserStore();

const selectedKeys = ref<string[]>([]);
const openedKeys = ref<string[]>([]);
const handleItemClick = (key: string) => {
  if (key === '-1') {
    router.push('/index')
  } else {
    const findRoute = userStore.menuRouteMap.get(key)
    if (findRoute) {
      router.push(findRoute.path)
    } else {
      errorMessage('菜单路由不存在');
    }
  }
}
const updateSelectedKey = () => {
  const menuId = route.name ? route.name.toString() : undefined;
  if (menuId) {
    if (menuId === 'index') {
      selectedKeys.value = ['-1'];
      return;
    }
    selectedKeys.value = [menuId];
    const parentIds = userStore.getMenuParentIdListMap().get(menuId)
    openedKeys.value = parentIds || [];
  }
}

watch(route, () => {
  updateSelectedKey();
})

onMounted(() => {
  updateSelectedKey();
})
</script>

<template>
  <div class="side-div">
    <a-menu :style="{ height: '100%', width: configStore.sideWidth }"
            show-collapse-button
            @menu-item-click="handleItemClick"
            v-model:selected-keys="selectedKeys"
            v-model:open-keys="openedKeys"
            accordion
    >
      <a-menu-item :key="'-1'">
        <template #icon><icon-home /></template>
        首页
      </a-menu-item>
      <template v-for="item in userStore.treeMenu" :key="item.menuId">
        <template v-if="item.isVisible">
          <template v-if="item.menuType === 1 || item.children">
            <a-sub-menu :key="item.menuId.toString()">
              <template #icon>
                <component :is="ArcoIcons[item.icon as keyof typeof ArcoIcons]" />
              </template>
              <template #title>{{ item.menuName }}</template>
              <template v-for="(menu, index) in item.children" :key="index">
                <a-menu-item v-if="menu.isVisible" :key="menu.menuId.toString()" :disabled="menu.isDisable">
                  <template #icon>
                    <component :is="ArcoIcons[menu.icon as keyof typeof ArcoIcons]" />
                  </template>
                  {{ menu.menuName }}
                </a-menu-item>
              </template>
            </a-sub-menu>
          </template>
          <template v-else>
            <a-menu-item :key="item.menuId.toString()">
              <template #icon>
                <component :is="ArcoIcons[item.icon as keyof typeof ArcoIcons]" />
              </template>
              {{ item.menuName }}
            </a-menu-item>
          </template>
        </template>
      </template>
    </a-menu>
  </div>
</template>

<style scoped lang="scss">
.side-div {
  box-shadow: 8px 0 8px -8px var(--shadow-color);
  height: 100%;
  border-top: 1px solid var(--border-color);
  .logo {
    padding: 24px 0;
  }
}
</style>