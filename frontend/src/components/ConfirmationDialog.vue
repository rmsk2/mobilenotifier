<script>
import PopupModal from './PopupModal.vue'

export default {
  components: { PopupModal },
  data() {
    return {
      title: "",
      message: "",
      okButton: "",
      cancelButton: "",

      resolvePromise: undefined,
      rejectPromise: undefined,
    }
  },
  methods: {
    show(title, message, okButton, cancelButton = "Abbrechen") {
      this.title = title
      this.message = message
      this.okButton = okButton
      this.cancelButton = cancelButton
      
      this.$refs.popup.open()
      
      return new Promise((resolve, reject) => {
        this.resolvePromise = resolve
        this.rejectPromise = reject
      })
    },
    confirm() {
      this.$refs.popup.close()
      this.resolvePromise(true)
    },
    cancel() {
      this.$refs.popup.close()
      this.resolvePromise(false)
    },
  },
}
</script>

<template>
  <popup-modal ref="popup">
    <h3 style="margin-top: 0" class="conf-header">{{ title }}</h3>
    <p>{{ message }}</p>
    <div class="confirm-btns">
      <button @click="cancel">{{ cancelButton }}</button>
      <button @click="confirm">{{ okButton }}</button>
    </div>
  </popup-modal>
</template>

<style scoped>
.confirm-btns {
  display: flex;
  flex-direction: row;
  justify-content: space-between;
}
</style>