<script>
export default {
  data() {
    return {
      count: 0
    }
  },
  props: ['reminders']
}
</script>

<template>
  <div id="div-allentries" class="work-entry">
    <h1>Übersicht über alle Ereignisse</h1>
      <li v-for="item in reminders">
        {{ item.description }}
      </li>    
  </div>
</template>