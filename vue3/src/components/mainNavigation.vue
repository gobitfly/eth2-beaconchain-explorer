<template>
  <nav id="nav" class="main-navigation navbar navbar-expand-lg navbar-light">
    <div class="container">
      <router-link class="navbar-brand" to="/">
        <brandSVG></brandSVG>
        <span class="brand-text">{{ SiteBrand }}</span>
      </router-link>
      <button class="navbar-toggler" type="button" data-toggle="collapse" data-target="#navbarSupportedContent" aria-controls="navbarSupportedContent" aria-expanded="false" aria-label="Toggle navigation">
        <span class="navbar-toggler-icon"></span>
      </button>
      <div class="collapse navbar-collapse" id="navbarSupportedContent">
        <ul class="navbar-nav ml-auto">
          <template v-for="menuItem in pageData.MainMenuItems">
            <li class="nav-item" :class="{ active: menuItem.active, dropdown: menuItem.Groups }">
              <template v-if="menuItem.Groups">
                <a class="nav-link dropdown-toggle" href="#" :id="menuItem.Label" role="button" data-toggle="dropdown" aria-haspopup="true" aria-expanded="false">
                  <span class="nav-text">{{ menuItem.Label }}</span>
                </a>
                <div :class="menuItem.HasBigGroups ? 'dropdown-menu dropdown-menu-right p-sm-2 p-md-4 px-lg-2 align-content-between flex-wrap flex-lg-nowrap' : 'dropdown-menu'" :aria-labelledby="'navbarDropdown' + menuItem.Label">
                  <template v-if="menuItem.HasBigGroups">
                    <template v-for="group in menuItem.Groups">
                      <div class="mx-lg-2 mt-2" style="flex: 1 1 240px">
                        <span class="ml-4" style="font-size: 18px; font-weight: 700; letter-spacing: 0.3px">{{ menuItem.Label }}</span>
                        <mainNavigationItem v-bind="link" v-for="link in group.Links"></mainNavigationItem>
                      </div>
                    </template>
                  </template>
                  <template v-else>
                    <mainNavigationItem v-bind="menuItem"></mainNavigationItem>
                    <hr />
                  </template>
                </div>
              </template>
              <template v-else>
                <router-link class="nav-link" :to="menuItem.Path"> {{ menuItem.Label }}</router-link>
              </template>
            </li>
            <!--
         
          {{ range . }}
            {{ $hasBigGroups := .HasBigGroups }}
            {{ $numberOfGroups := (len .Groups) }}
            <li class="nav-item {{ if .IsActive }}active{{ end }} {{ if $numberOfGroups }}dropdown{{ end }}">
              {{ if len .Groups }}
                <a class="nav-link dropdown-toggle" href="#" id="navbarDropdown{{ .Label }}" role="button" data-toggle="dropdown" aria-haspopup="true" aria-expanded="false">
                  <span class="nav-text">{{ .Label }}</span>
                </a>

                <div class="dropdown-menu {{ if $hasBigGroups }}dropdown-menu-right p-sm-2 p-md-4 px-lg-2 align-content-between flex-wrap flex-lg-nowrap{{ end }}" aria-labelledby="navbarDropdown{ .Label }}">
                  {{ range $groupIndex, $group := .Groups }}
                    {{ if $hasBigGroups }}
                      <div class="mx-lg-2 mt-2" style="flex: 1 1 240px;">
                        <span class="ml-4" style="font-size: 18px; font-weight: 700; letter-spacing: .3px;">{{ .Label }}</span>
                        {{ range .Links }}
                          {{ template "mainNavigationItem" . }}
                        {{ end }}
                      </div>
                    {{ else }}
                      {{ $numberOfLinks := (len .Links) }}
                      {{ range $index, $link := .Links }}
                        {{ template "mainNavigationItem"  $link }}
                        {{ if and (eq $index (sub $numberOfLinks 1)) (not (eq $groupIndex (sub $numberOfGroups 1))) }}
                          <hr />
                        {{ end }}
                      {{ end }}
                    {{ end }}
                  {{ end }}
                </div>
              {{ else }}
                <a class="nav-link" href="{{ .Path }}">
                  <span class="nav-text">{{ .Label }}</span>
                </a>
              {{ end }}
            </li>
          {{ end }}
          -->
          </template>
        </ul>
      </div>
    </div>
  </nav>
</template>

<script setup>
import { onServerPrefetch, onMounted, computed } from "vue"
import { useState } from "@/stores/state.js"

// https://www.npmjs.com/package/vite-svg-loader
import brandSVG from "#/static/img/brand.svg"

import mainNavigationItem from "@/components/mainNavigationItem.vue"

const state = useState()

//config.Frontend.SiteBrand
const SiteBrand = "beaconcha.in"
const pageData = computed(() => state.pageData)
</script>
