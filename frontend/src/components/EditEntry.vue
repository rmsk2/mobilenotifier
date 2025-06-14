<script>
import { reminderAnniversary, reminderOneShot } from './reminderapi';
import { warningMorningBefore, warningNoonBefore, warningEveningBefore, warningWeekBefore, warningSameDay } from './reminderapi';


export default {
  data() {
    return {
    }
  },
  props: ['data', 'reminderid', 'api', 'isnew', 'recipients'],
  emits: ['error-occurred'],
  methods: {
    saveData() {
      this.$emit('error-occurred', "Daten gespeichert")
    }
  }
}
</script>

<template>
  <div id="div-edit-entry" class="work-entry">
    <h2 v-if="isnew">Ereignis erstellen</h2>
    <h2 v-else>Ereignis bearbeiten</h2>
    <button @click="saveData">Daten speichern</button>
  </div>
</template>