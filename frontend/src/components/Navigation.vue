<script>

import { monthSelected, newSelected, allSelected, aboutSelected } from './globals';

export default {
  data() {
    return {
      monthSelected: monthSelected,
      newSelected: newSelected,
      allSelected: allSelected,
      aboutSelected: aboutSelected,

      idPressed: monthSelected
    }
  },  
  props: ['currentstate'],
  emits: ['select-nav'],
  watch: {
    currentstate: {
        handler(newVal){
          this.idPressed = newVal
        },
        immediate: true
    }
  },  
  methods: {
    emitEvent(id) {
      this.idPressed = id;
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
    <button id="nav_nmonth" class="navbutton" :class="{active: idPressed === monthSelected}" @click="emitMonth">Ereignisse im gewählten Monat anzeigen</button>
    <button id="nav_new" class="navbutton" :class="{active: idPressed === newSelected}" @click="emitNew">Ereignis anlegen/bearbeiten</button>
    <button id="nav_all" class="navbutton" :class="{active: idPressed === allSelected}" @click="emitAll">Alle Ereignisse anzeigen</button>
    <button id="nav_about" class="navbutton" :class="{active: idPressed === aboutSelected}" @click="emitAbout">Über diese Anwendung</button>  
  </div>
</template>