<template>
  <div class="info-banner-container">
    <div class="info-banner-content container">
      <div id="banner-stats" class="info-banner-left">
        <template v-if="pageData.ShowSyncingMessage">
          <a data-toggle="tooltip" title="The explorer is currently syncing with the network" id="banner-status" style="white-space: nowrap" class="mr-2" href="/"><i class="fas fa-sync"></i> <span>|</span></a>
        </template>
        <template v-else>
          <router-link to="/" id="banner-home" style="white-space: nowrap" class="mr-2"><i class="fas fa-home"></i> <span>|</span></router-link>
        </template>
        <div data-toggle="tooltip" title="" data-original-title="Epoch" id="banner-epoch" class="info-item d-flex mr-2 mr-lg-3">
          <div class="info-item-header mr-1">
            <span class="item-icon"><i class="fas fa-history"></i></span>
            <span class="d-none d-sm-inline item-text">Ep<span class="d-none d-xl-inline">och</span></span>
          </div>
          <div class="info-item-body">
            <router-link :to="'/epoch/' + pageData.CurrentEpoch" id="banner-epoch-data"> {{ pageData.CurrentEpoch }} </router-link>
          </div>
        </div>
        <div data-toggle="tooltip" title="" data-original-title="Slot" class="d-none d-lg-block">
          <div id="banner-slot" class="info-item d-flex mr-2 mr-lg-3">
            <div class="info-item-header mr-1">
              <span class="item-icon"><i class="fas fa-cubes"></i></span>
              <span class="item-text">Slot</span>
            </div>
            <div class="info-item-body">
              <router-link id="banner-slot-data" :to="'/slot/' + pageData.CurrentSlot"> {{ pageData.CurrentSlot }}</router-link>
            </div>
          </div>
        </div>
        <template v-if="pageData.Mainnet">
          <div data-toggle="tooltip" title="" data-original-title="Price: 1 {{pageData.Rates.mainCurrencySymbol }} = {{ pageData.Rates.mainCurrencyTickerPriceKFormatted }} {{ .Rates.tickerCurrencySymbol }}" id="banner-eth-price">
            <div class="info-item d-flex mr-2 mr-lg-3">
              <div class="info-item-header mr-1">
                <span class="item-icon"><i class="fas fa-cubes"></i></span>
                <span class="d-none d-xl-inline item-text">Price</span>
              </div>
              <div class="info-item-body">
                <a id="banner-eth-price-data">
                  <span id="currentCurrencySymbol" class="currency-symbol">{{ pageData.Rates.tickerCurrencySymbol }}</span>
                  <span id="currentKFormattedPrice" class="k-formatted-price">{{ pageData.Rates.mainCurrencyTickerPriceKFormatted }}</span>
                  <span id="currentCurrencyPrice" class="price">{{ pageData.Rates.mainCurrencyTickerPriceFormatted }}</span>
                </a>
              </div>
            </div>
          </div>
        </template>
        <template v-else>
          <template v-if="pageData.GasNow">
            <div data-toggle="tooltip" title="" data-original-title="Gas Price" class="d-none d-lg-block">
              <div id="banner-slot" class="info-item d-flex mr-2 mr-lg-3">
                <div class="info-item-body">
                  <router-link id="banner-gpo-data" to="/gasnow"><i class="fas fa-gas-pump mr-1"></i>{{ pageData.GasNow.data.fast }}</router-link>
                </div>
              </div>
            </div>
          </template>
        </template>
        <template v-if="!pageData.ShowSyncingMessage && pageData.FinalizationDelay > 5">
          <div data-toggle="tooltip" title="" data-original-title="Finality: The last finalized epoch was {{ pageData.FinalizationDelay }} epochs ago" id="banner-fin" class="info-item d-flex mr-2 mr-lg-3">
            <div class="info-item-header mr-1">
              <span class="item-icon"><i class="fas fa-exclamation-triangle"></i></span>
            </div>
            <div class="info-item-body text-warning">
              <span id="banner-fin-data">{{ pageData.FinalizationDelay }}</span>
              <i class="fas fa-exclamation-triangle item-text"></i>
            </div>
          </div>
        </template>
      </div>
      <!-- Fill content -->
      <!-- <div class="d-flex flex-fill"></div> -->
      <div class="info-banner-center">
        <div class="info-banner-search">
          <div class="search-container">
            <form class="form-inline" action="/search" method="POST">
              <input id="banner-search-input" class="typeahead" autocomplete="off" name="search" type="text" placeholder="Public Key / Block Number / Block Hash / Graffiti / State Hash" aria-label="Search" />
            </form>
          </div>
          <a class="search-button"><i id="banner-search" class="fas fa-search"></i></a>
        </div>
      </div>
      <div class="info-banner-right">
        <template v-if="pageData.Mainnet">
          <div class="dropdown">
            <a class="btn btn-transparent btn-sm dropdown-toggle currency-dropdown-toggle" data-toggle="dropdown" aria-haspopup="true" aria-expanded="false">
              <div id="currencyFlagDropdown">
                <img class="currency-flag-dropdown-image" :src="`/img/${pageData.Rates.selectedCurrency}.svg`" />
              </div>
              <div id="currencyDropdown">{{ pageData.Rates.selectedCurrency }}</div>
            </a>
            <!--
          <div class="dropdown-menu dropdown-menu-right" aria-labelledby="currencyDropdown">
            {{ range .AvailableCurrencies }}
              <a tabindex="1" class="dropdown-item cursor-pointer" onClick="updateCurrency({{ . }})">
                <img class="currency-flag-option" src="/img/{{ . }}.svg" />
                <span class="currency-name">{{ getCurrencyLabel . }}</span>
                <span class="currency-symbol">{{ . }}</span>
              </a>
            {{ end }}
          </div>
          --></div>
        </template>

        <template v-if="pageData.User.authenticated">
          <div class="dropdown">
            <a class="btn btn-transparent btn-sm dropdown-toggle" id="userDropdown" data-toggle="dropdown" aria-haspopup="true" aria-expanded="false">
              <i class="fas fa-user-circle m-0 p-0"></i>
            </a>
            <div class="dropdown-menu dropdown-menu-right" aria-labelledby="userDropdown">
              <a class="dropdown-item" href="/user/notifications">Notifications</a>
              <a class="dropdown-item" href="/user/settings">Settings</a>
              <template v-if="pageData.User.user_group === 'ADMIN'">
                <a class="dropdown-item" href="/user/global_notification">Global Notification</a>
                <a class="dropdown-item" href="/user/ad_configuration">Ad Configuration</a>
                <a class="dropdown-item" href="/user/explorer_configuration">Explorer Configuration</a>
              </template>
              <a data-no-instant class="dropdown-item" href="/logout">Logout</a>
            </div>
          </div>
        </template>

        <template v-else>
          <div class="d-md-none">
            <div class="dropdown d-md-none">
              <a class="btn btn-transparent btn-sm dropdown-toggle" id="loginDropdown" data-toggle="dropdown" aria-haspopup="true" aria-expanded="false">
                <i style="margin: 0; padding: 0" class="fas fa-sign-in-alt"></i>
              </a>
              <div class="dropdown-menu dropdown-menu-right" aria-labelledby="loginDropdown">
                <a class="dropdown-item" href="/login">Log in</a>
                <a class="dropdown-item" href="/register">Sign Up</a>
              </div>
            </div>
          </div>
          <a href="/login" class="mr-3 d-none d-md-flex"><span>Log in</span></a>
          <a href="/register" class="btn btn-primary btn-sm d-none d-md-flex"
            ><span class="text-white"><b>Sign Up</b></span></a
          >
        </template>
      </div>
    </div>
  </div>
</template>

<script setup>
import { onServerPrefetch, onMounted, computed } from "vue"
import { useState } from "@/stores/state.js"

const state = useState()

const epoch = computed(() => state.indexData.epoch)
const pageData = computed(() => state.pageData)
</script>

// stylus example

<style scoped lang="stylus">
#banner-status,#banner-home
  white-space nowrap
</style>
