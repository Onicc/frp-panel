<template>
  <Teleport to="body">
    <Transition name="modal">
      <div v-if="show" class="modal-backdrop" role="dialog" aria-modal="true" @click.self="closeOnClickOutside && emit('close')">
        <section class="modal" :class="widthClass" :aria-label="title" @click.stop>
          <header class="modal-header"><h2>{{ title }}</h2><button v-if="showCloseButton" class="icon-button" type="button" :aria-label="t('common.close')" @click="emit('close')"><Icon name="x" /></button></header>
          <div class="modal-body"><slot /></div>
          <footer v-if="$slots.footer" class="modal-actions"><slot name="footer" /></footer>
        </section>
      </div>
    </Transition>
  </Teleport>
</template>
<script setup lang="ts">
import { computed, onBeforeUnmount, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import Icon from '../icons/Icon.vue'
const props = withDefaults(defineProps<{ show: boolean; title: string; width?: 'narrow'|'normal'|'wide'; closeOnClickOutside?: boolean; showCloseButton?: boolean }>(), { width: 'normal', closeOnClickOutside: false, showCloseButton: true })
const emit = defineEmits<{ close: [] }>()
const { t } = useI18n()
const widthClass = computed(() => ({ narrow: 'modal-narrow', normal: '', wide: 'modal-wide' }[props.width]))
const onKeydown = (event: KeyboardEvent) => { if (event.key === 'Escape' && props.show) emit('close') }
watch(() => props.show, open => { document.body.classList.toggle('modal-open', open); if (open) document.addEventListener('keydown', onKeydown); else document.removeEventListener('keydown', onKeydown) }, { immediate: true })
onBeforeUnmount(() => { document.body.classList.remove('modal-open'); document.removeEventListener('keydown', onKeydown) })
</script>
