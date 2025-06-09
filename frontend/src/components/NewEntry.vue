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
  <h1>New Entry</h1>
  <EditEntry></EditEntry>
</template>