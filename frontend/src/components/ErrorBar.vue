<script>
export default {
  data() {
    return {
      timerId: null
    }
  },
  props: ['usermessage', 'interval'],
  emits: ['reset-error'],
  watch: {
    usermessage(newVal) {
      this.stopTimer()

      if (newVal == "") {
        return;
      }

      this.timerId = setTimeout(() => {
        this.signalReset()
        }, this.interval)
      }
    },
  methods: {
    signalReset() {
      this.stopTimer()
      this.$emit('reset-error');
    },
    stopTimer() {
      if (this.timerId != null) {
        clearTimeout(this.timerId);
        this.timerId = null;
      }
    }
  }
}
</script>

<template>
  <div id="div-error-bar" class="error-bar">
    <span class="message-text">{{ usermessage }}</span>
  </div>
</template>