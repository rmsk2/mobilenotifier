<script>
import AllEntries from './components/AllEntries.vue'
import MonthlyEntries from './components/MonthlyEntries.vue'
import Navigation from './components/Navigation.vue'
import NewEntry from './components/NewEntry.vue'
import { monthSelected, newSelected, allSelected } from './components/globals';
import { ReminderAPI } from './components/reminderapi';

export default {
  data() {
    return {
      apiUrlBase: import.meta.env.VITE_API_URL,
      accessToken: "egal",
      showMonthly: true,
      showAll: false,
      showNew: false,
      monthlyEntries: [{message: "Eins"}, {message: "Zwei"}],
      result: "",
      allRecipients: []
    }
  },
  methods: {
    async sendSms2() {
      let api = new ReminderAPI(this.apiUrlBase, this.accessToken)

      let res = await api.sendSms("Dies ist ein Test", "martin")
      if (res.error) {
        this.result = res.data
      } else {
        this.result = "Success"
      }
    },
    async getRecipients() {
      let api = new ReminderAPI(this.apiUrlBase, this.accessToken)

      let res = await api.getRecipients()
      if (res.error) {
        return
      }

      this.allRecipients = res.data
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
  <button @click="sendSms2">Testnachricht senden</button>
  <button @click="getRecipients">Alle möglichen Empfänger abrufen</button>
  <p/>
  <li v-for="r in allRecipients">
    {{ r }}
  </li>
  <p/>
  {{ result }}
  <section class="work-items">
    <AllEntries v-if="showAll"></AllEntries>
    <MonthlyEntries :reminders="monthlyEntries" v-if="showMonthly"></MonthlyEntries>
    <NewEntry v-if="showNew"></NewEntry>
  </section>
</template>