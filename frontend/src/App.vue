<script>
export default {
  data() {
    return {
      count: 0,
      result: "",
      apiResult: "",
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
    },
    testApi() {
      this.apiResult = "Trying ..."
      fetch(this.apiUrlBase + "test")
        .then(response => {
          if (!response.ok) {
            return Promise.reject(response);
          }
          return response.text()
        })
        .then(data => this.apiResult = data)
        .catch(error =>  this.apiResult = "Failure")
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
  <p/>
  <button @click="testApi">Test API</button>
  <p/>
  <div> {{ apiResult }}</div>  
</template>