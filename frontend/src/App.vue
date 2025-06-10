<script>
import AllEntries from './components/AllEntries.vue'
import MonthlyEntries from './components/MonthlyEntries.vue'
import Navigation from './components/Navigation.vue'
import NewEntry from './components/NewEntry.vue'
import { monthSelected, newSelected, allSelected } from './components/globals';

export default {
  data() {
    return {
      apiUrlBase: import.meta.env.VITE_API_URL,
      showMonthly: true,
      showAll: false,
      showNew: false,
      monthlyEntries: [{message: "Eins"}, {message: "Zwei"}],
      message: "Dies ist ein Test",
      result: ""
    }
  },
  methods: {
    sendSms() {
      this.result = "Sending ..."
      fetch(this.apiUrlBase + "send/martin", {
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
    },
    makeAllComponentsInvisible() {
      this.showAll = false;
      this.showMonthly = false;
      this.showNew = false;
    },
    add() {
      this.monthlyEntries.push({message: "Noch eins ..."})
    },
    showComponents(value) {
      this.makeAllComponentsInvisible();

      if (value === monthSelected) {
        this.showMonthly = true;
      }

      if (value === allSelected) {
        this.showAll = true;
      }
      
      if (value === newSelected) {
        this.showNew = true;
      }      
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
  <section class="section-navitems">
    <Navigation @select-nav="showComponents"></Navigation>
  </section>
  <button @click="add">Testweise hinzufügen</button>
  <button @click="sendSms">Testnachricht senden</button>
  <p/>
  {{ result }}
  <section class="work-items">
    <AllEntries v-if="showAll"></AllEntries>
    <MonthlyEntries :reminders="monthlyEntries" v-if="showMonthly"></MonthlyEntries>
    <NewEntry v-if="showNew"></NewEntry>
  </section>
</template>