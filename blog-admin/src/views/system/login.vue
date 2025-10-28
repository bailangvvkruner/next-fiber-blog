<script setup lang="ts">
import { useRoute, useRouter } from 'vue-router'
import { onMounted, reactive, ref } from 'vue'
import { useLocalStorage } from '@/hooks/local-storage'
import { useArcoMessage } from '@/hooks/message'
import type { LoginForm } from '@/api/system/user/types'
import { userApi } from '@/api/system/user'
import { useLocalUserStore } from '@/stores/user'
import { LOCAl_USER_KEY, TOKEN_KEY } from '@/assets/script/constant'
import { menuApi } from '@/api/system/menu'
import { buildRoute } from '@/router'

const { get, set, remove } = useLocalStorage();
const { successMessage } = useArcoMessage();
const route = useRoute();
const router = useRouter();

const SAVE_DATA_KEY = "login_data";

const isSavePassword = ref<boolean>(false);

const defaultFormData: LoginForm = {
  username: undefined,
  password: undefined,
}
const formData = reactive<LoginForm>({ ...defaultFormData })
const loginButtonLoading = ref<boolean>(false);
const handleSubmit = async () => {
  loginButtonLoading.value = true;
  const passwordBase64 = btoa(formData.password || '');
  const form: LoginForm = {
    username: formData.username,
    password: passwordBase64,
  }
  try {
    const result = await userApi.login(form)
    const { code, data } = result;
    if (code !== 200 || !data) {
      return;
    }
    // 登录成功保存Token
    useLocalStorage().set(TOKEN_KEY, data, undefined);
    if (isSavePassword.value) {
      const saveLoginData = btoa(JSON.stringify(formData));
      set<string>(SAVE_DATA_KEY, saveLoginData, undefined)
    } else {
      remove(SAVE_DATA_KEY);
    }
    // 获取登录用户详情和用户菜单
    Promise.all([userApi.queryUserInfo(), menuApi.listTreeMenu()]).then(values => {
      const [userResult, menuResult] = values;
      // 保存用户信息
      useLocalStorage().set(LOCAl_USER_KEY, userResult.data, undefined);
      useLocalUserStore().userInfo = userResult.data;
      // 构建路由并保存路由信息
      const routeList = buildRoute(menuResult.data);
      useLocalUserStore().setTreeMenu(menuResult.data);
      useLocalUserStore().setMenuRoute(routeList);
      successMessage('登录成功')
      if (route.query && route.query.redirect) {
        router.push({ path: route.query.redirect as string });
      } else {
        router.push({ path: '/index' })
      }
    })
  } finally {
    loginButtonLoading.value = false;
  }
}

const readerSavaLoginData = () => {
  const base64Data = get<string>(SAVE_DATA_KEY);
  if (base64Data) {
    isSavePassword.value = true;
    const { username, password } = JSON.parse(atob(base64Data)) as LoginForm;
    formData.username = username;
    formData.password = password;
  }
}

onMounted(() => {
  readerSavaLoginData();
})
</script>

<template>
  <div class="login-container flex">
    <div class="left-background"></div>
    <div class="right-div flex justify-center item-center">
      <div ref="loginCardRef" class="login-card radius-xl">
        <h1 class="title">欢迎访问博客后台管理系统</h1>
        <a-form :model="formData" layout="vertical" @submit-success="handleSubmit">
          <a-form-item field="username" label="用户名" hide-label :rules="[{required: true, message: '用户名不能为空'}]">
            <a-input v-model="formData.username" size="large" placeholder="请输入用户名">
              <template #prefix>
                <icon-user />
              </template>
            </a-input>
          </a-form-item>
          <a-form-item field="password" label="密码" hide-label :rules="[{required: true, message: '密码不能为空'}]">
            <a-input-password v-model="formData.password" size="large" placeholder="请输入密码">
              <template #prefix>
                <icon-unlock />
              </template>
            </a-input-password>
          </a-form-item>
          <a-form-item>
            <a-button type="primary" html-type="submit" size="large" long :loading="loginButtonLoading">登录</a-button>
          </a-form-item>
        </a-form>
        <div class="save-password">
          <input id="save-password" type="checkbox" v-model="isSavePassword" />
          <label for="save-password">
            保存密码
          </label>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped lang="scss">
.login-container {
  height: 100vh;
  width: 100%;
  background-color: var(--background-color);
  .left-background {
    width: 500px;
    flex-shrink: 1;
    background-image: url("/images/login-bg.png");
    background-repeat: no-repeat;
    background-size: 80%;
    background-position: center;
    background-color: #87CEFA;
  }
  .right-div {
    flex: 1;
    .login-card {
      position: relative;
      overflow: hidden;
      row-gap: var(--space-sm);
      padding: var(--space-xxl);
      box-shadow: 0 0 8px var(--shadow-color);
      background-color: transparent;
      z-index: 2;
      transition: all 300ms ease;
      &:hover {
        box-shadow: 0 0 16px var(--shadow-color);
      }
      &::before {
        content: "";
        position: absolute;
        left: 0;
        top: 0;
        bottom: 0;
        right: 0;
        background-color: var(--card-color);
        z-index: -3;
      }
      .title {
        font-weight: 600;
        padding: var(--space-sm) 0 var(--space-xl) 0;
      }
      .code-input {
        column-gap: var(--space-md);
        .verify-code {
          flex-shrink: 1;
          width: 100px;
          background-color: #1E90FF;
        }
      }
      .save-password {
        font-size: 14px;
      }
    }
  }
}
.login-container, .right-div, .login-card {
  transition: padding 300ms ease;
}
@media (max-width: 950px) {
  .login-container {
    display: block;
    position: relative;
    .left-background {
      position: absolute;
      left: 0;
      top: 0;
      width: 100%;
      height: 100%;
      z-index: 1;
    }
    .right-div {
      position: absolute;
      left: 0;
      top: 0;
      width: 100%;
      height: 100%;
      z-index: 2;
    }
  }
}
@media (max-width: 550px) {
  .login-container {
    .right-div {
      padding: var(--space-md);
      .login-card {
        padding: var(--space-md);
      }
    }
  }
}
</style>