<template><article class="stat-card"><div class="stat-heading"><span>{{ label }}</span><Icon :name="icon" size="sm" /></div><strong>{{ value }}</strong><small v-if="hint">{{ hint }}</small></article></template>
<script setup lang="ts">import Icon from '../icons/Icon.vue'; withDefaults(defineProps<{label:string;value:string|number;hint?:string;icon?:string}>(),{icon:'chart'})</script>
