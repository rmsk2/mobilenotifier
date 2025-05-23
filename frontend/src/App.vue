<script>
export default {
  data() {
    return {
      count: 0,
      result: "",
      apiUrlBase: import.meta.env.VITE_API_URL,
      message: "",
      recipient: "martin"
    }
  },
  methods: {
    sendSms() {
      this.result = "Sending ..."
      fetch(this.apiUrlBase + "send/" + this.recipient, {
          method: "post",
          headers: {
            'Accept': 'application/json',
            'Content-Type': 'application/json'
          },
          body: JSON.stringify({
            message: this.message
          })
        })
        .then(response => {
          if (response.ok) {
            this.result = "Success"
          } else {
            this.result = `Failure (${response.status})`
          }
        })
        .catch(error =>  this.result = "Failure")
    }
  }
}

</script>

<template>
  <textarea v-model="message" placeholder="Enter message"></textarea >
  <p/>
  <button @click="sendSms">Send SMS</button>
  <p/>
  <div> {{ result }}</div>  

</template>