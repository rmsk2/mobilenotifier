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
    <h1>Monthly entries</h1>
      <li v-for="item in reminders">
        {{ item.message }}
      </li>
  </div>
</template>