<script>

import { monthSelected, newSelected, allSelected, aboutSelected } from './globals';

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
    },
    emitAbout() {
      this.emitEvent(aboutSelected)
    }    
  },  
}
</script>

<template>
  <div class="navbar"> 
    <button id="nav_nmonth" class="navbutton" @click="emitMonth">Alle Ereignisse im gewählten Monat anzeigen</button>
    <button id="nav_new" class="navbutton" @click="emitNew">Neues Ereignis anlegen</button>
    <button id="nav_all" class="navbutton" @click="emitAll">Alle Ereignisse anzeigen</button>
    <button id="nav_about" class="navbutton" @click="emitAbout">Über diese Anwendung</button>  
  </div>
</template>