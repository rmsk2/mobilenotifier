<script>
export default {
  data() {
    return {
    }
  },
  props: ['reminders']
}
</script>

<template>
  <div id="div-monthly" class="work-entry">
    <h1>Ereignisse im gewählten Monat</h1>
      <li v-for="item in reminders">
        {{ item.reminder.description }}, {{ item.next_occurrance }}
      </li>     
  </div>
</template>