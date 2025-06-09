<script>
import EditEntry from './EditEntry.vue';

export default {
  data() {
    return {
      count: 0
    }
  },
  components: {
    EditEntry,
  }    
}
</script>

<template>
  <div id="div-new" class="work-entry">
    <h1>New Entry</h1>
    <EditEntry></EditEntry>
  </div>
</template>