<template>
  <router-link :to="Path" class="dropdown-item" :style="[IsHighlighted ? { padding: '0 0.625rem;' } : {}]">
    <span :class="[IsHighlighted ? ['btn btn-primary d-flex align-items-center'] : []]">
      <span class="nav-icon">
        <component v-if="CustomIcon" :is="CustomIcon"></component>
        <i :class="'fas ' + Icon + ' mr-2'"></i>
      </span>
      <span :class="[IsHighlighted ? ['text-white'] : []]"> {{ Label }}</span>
    </span>
  </router-link>
</template>

<script setup>
// todo
// dynamic import of svg wont replace <componenttag> with svg
// https://github.com/jpkleemans/vite-svg-loader/issues/85

import { defineProps } from "vue"

import ethermine_staking_logo_svg from "#/static/img/ethermine_staking_logo.svg?url"
import ethermine_stake_logo_svg from "#/static/img/ethermine_stake_logo.svg?url"
import eversteel_logo_svg from "#/static/img/eversteel_logo.svg"
import webhook_logo_svg from "#/static/img/webhook_logo.svg"

const props = defineProps(["Label", "Path", "CustomIcon", "Icon", "IsHidden", "IsHighlighted"])
</script>
