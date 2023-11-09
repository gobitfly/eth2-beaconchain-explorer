<template>
  <template v-if="isMounted"><slot /></template>
  <template v-else><slot name="placeholder" /></template>
</template>

<script setup>
import { ref, onMounted } from "vue"
const isMounted = ref(false)
onMounted(() => {
  isMounted.value = true
})
</script>
