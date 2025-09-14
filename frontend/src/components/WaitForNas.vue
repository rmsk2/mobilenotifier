<script>
import PopupModal from './PopupModal.vue'

export default {
  components: { PopupModal },
  data() {
    return {
      title: "",
      message: "",
      timerId: null,
      interval: 1000
    }
  },
  methods: {
    show(title, message) {
      this.hide()

      this.title = title
      this.message = message
      
      this.timerId = setInterval(() => { this.$refs.popup.open() }, this.interval)
    },      
    hide() {
      this.stopTimer()
      this.$refs.popup.close()
    },
    stopTimer() {
      if (this.timerId != null) {
        clearInterval(this.timerId);
        this.timerId = null;
      }
    }    
  }
}
</script>

<template>
  <popup-modal ref="popup">
    <h3 style="margin-top: 0" class="conf-header">{{ title }}</h3>
    <p>{{ message }}</p>
  </popup-modal>
</template>
