<script>

import { monthSelected, newSelected, allSelected } from './globals';

export default {
  emits: ['select-nav'],
  methods: {
    emitEvent(id) {
      this.$emit('select-nav', id);
    },
    emitNew() {
      this.emitEvent(newSelected)
    },
    emitMonth() {
      this.emitEvent(monthSelected)
    },
    emitAll() {
      this.emitEvent(allSelected)
    }
  },  
}
</script>

<template>
  <div class="navbar"> 
    <button id="nav_nmonth" class="navbutton" @click="emitMonth">Alle Ereignisse im aktuellen Monat anzeigen</button>
    <button id="nav_new" class="navbutton" @click="emitNew">Neues Ereignis anlegen</button>
    <button id="nav_all" class="navbutton" @click="emitAll">Alle Ereignisse anzeigen</button>  
  </div>
</template>