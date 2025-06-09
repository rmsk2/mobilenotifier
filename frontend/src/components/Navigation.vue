<script>

import { monthSelected, newSelected, allSelected } from './globals';

export default {
  emits: ['select-nav'],
  methods: {
    emitEvents(id) {
      switch (id) {
        case 10:
          this.$emit('select-nav', monthSelected);
          break;
        case 11:
          this.$emit('select-nav', newSelected);
          break;
        default:
          this.$emit('select-nav', allSelected);
      }      
    }
  },  
}
</script>

<template>  
  <button id="nav_nmonth" class="navbutton" @click="emitEvents(10)">Alle Ereignisse im aktuellen Monat anzeigen</button>
  <button id="nav_new" class="navbutton" @click="emitEvents(11)">Neues Ereignis anlegen</button>
  <button id="nav_all" class="navbutton" @click="emitEvents(12)">Alle Ereignisse anzeigen</button>  
</template>