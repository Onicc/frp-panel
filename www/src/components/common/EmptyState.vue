<template><div class="empty"><Icon :name="icon" size="lg" /><strong>{{ title }}</strong><span>{{ message }}</span><slot /></div></template>
<script setup lang="ts">import Icon from '../icons/Icon.vue'; withDefaults(defineProps<{ title: string; message?: string; icon?: string }>(), { message: '', icon: 'inbox' })</script>
