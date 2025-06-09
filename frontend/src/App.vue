<script>
import AllEntries from './components/AllEntries.vue'
import MonthlyEntries from './components/MonthlyEntries.vue'
import Navigation from './components/Navigation.vue'
import NewEntry from './components/NewEntry.vue'

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
            'Content-Type': 'application/json',
            'X-Token': 'egal'
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
  },
  components: {
    AllEntries,
    MonthlyEntries,
    Navigation,
    NewEntry
  }  
}

</script>

<template>
  <Navigation></Navigation>
  <AllEntries></AllEntries>
  <MonthlyEntries></MonthlyEntries>
  <NewEntry></NewEntry>
</template>