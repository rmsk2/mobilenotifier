<script>
import AllEntries from './components/AllEntries.vue'
import ErrorBar from './components/ErrorBar.vue';
import MonthlyEntries from './components/MonthlyEntries.vue'
import Navigation from './components/Navigation.vue'
import NewEntry from './components/NewEntry.vue'
import { monthSelected, newSelected, allSelected } from './components/globals';
import { ReminderAPI, ReminderData } from './components/reminderapi';
import { reminderAnniversary, reminderOneShot } from './components/reminderapi';
import { warningMorningBefore, warningNoonBefore, warningEveningBefore, warningWeekBefore, warningSameDay } from './components/reminderapi';

export default {
  data() {
    return {
      apiUrlBase: import.meta.env.VITE_API_URL,
      accessToken: "egal",
      showMonthly: true,
      showAll: false,
      showNew: false,
      overviewEntries: [],
      entriesInMonth: [],
      result: "",
      allRecipients: [],
      reminderId: "",
      reminderData: null
    }
  },
  methods: {
    async createNew() {
      let api = new ReminderAPI(this.apiUrlBase, this.accessToken)

      let t = new Date();
      t = new Date(t.getTime() + (2 * 60 * 60 * 1000));

      let r = new ReminderData(reminderOneShot, 0, [warningEveningBefore, warningSameDay], t, "Dies ist ein Test", this.allRecipients);

      let res = await api.createNewReminder(r);
      if (res.error) {
        this.setErrorMessage("Kann Eintrag nicht anlegen");
      }

      this.reminderData= res.data
    },
    async getReminder() {
      let api = new ReminderAPI(this.apiUrlBase, this.accessToken);

      let res = await api.readReminder(this.reminderId);
      this.reminderData = res;
    },
    async modifyReminder() {
      let api = new ReminderAPI(this.apiUrlBase, this.accessToken);

      let res = await api.readReminder(this.reminderId);
      if (res.error) {
        return
      }

      if (!res.data.found) {
        return
      }

      let oldData = res.data.data
      let id = oldData.id
      let remData = new ReminderData(oldData.kind, oldData.param + 1, oldData.warning_at, oldData.spec, oldData.description, oldData.recipients)

      res = await api.updateReminder(remData, id)
      this.reminderData = res;
    },
    async deleteReminder() {
      let api = new ReminderAPI(this.apiUrlBase, this.accessToken);

      let res = await api.deleteReminder(this.reminderId);
      this.reminderData = res;
    },
    async sendSms2() {
      let api = new ReminderAPI(this.apiUrlBase, this.accessToken)

      let res = await api.sendSms("Dies ist ein Test", "martin")
      if (res.error) {
        this.setErrorMessage("Versand fehlgeschlagen");
      }
    },
    async getRecipients() {
      let api = new ReminderAPI(this.apiUrlBase, this.accessToken)

      let res = await api.getRecipients()
      if (res.error) {
        this.setErrorMessage("Kann Empfängerliste nicht abrufen")
        this.allRecipients = [];
        return
      }

      this.allRecipients = res.data
    },
    resetErrors() {
      this.result = ""
    },
    setErrorMessage(msg) {
      this.result = msg
    },    
    makeAllComponentsInvisible() {
      this.showAll = false;
      this.showMonthly = false;
      this.showNew = false;
    },
    async getOverview() {
      let api = new ReminderAPI(this.apiUrlBase, this.accessToken)

      let res = await api.getOverview();      
      if (res.error) {
        this.setErrorMessage("Kann Übersicht nicht abrufen");
        this.overviewEntries = [];
        return;
      }

      this.overviewEntries = res.data;
    },
    async getEventsInMonth() {
      let now = new Date();
      let api = new ReminderAPI(this.apiUrlBase, this.accessToken)

      let m = now.getMonth() + 1
      let y = now.getFullYear()

      let res = await api.getEventsInMonth(m, y);
      
      if (res.error) {
        this.setErrorMessage("Kann Ereignisse nicht abrufen");
        this.entriesInMonth = [];
        return
      }
      
      this.entriesInMonth = res.data;      
    },
    async showComponents(value) {
      this.makeAllComponentsInvisible();
      this.result = ""

      if (value === monthSelected) {
        await this.getEventsInMonth()
        this.showMonthly = true;
      }

      if (value === allSelected) {
        await this.getOverview()        
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
    NewEntry,
    ErrorBar
  },
  async beforeMount() {
    await this.getRecipients()
    await this.showComponents(monthSelected)
  }
}

</script>

<template>
  <section class="section-navitems">
    <Navigation @select-nav="showComponents"></Navigation>
  </section>
  <input type="text" v-model="reminderId" size="40"></input>
  <p/>
  <button @click="getReminder">Ereignis abrufen</button>
  <button @click="deleteReminder">Ereignis löschen</button>
  <button @click="createNew">Neues Ereignis anlegen</button>
  <button @click="modifyReminder">Ereignis modifizieren</button>
  <p/>
  {{ reminderData }}
  <p/>
  <section class="work-items">
    <AllEntries :reminders="overviewEntries" v-if="showAll"></AllEntries>
    <MonthlyEntries :reminders="entriesInMonth" v-if="showMonthly"></MonthlyEntries>
    <NewEntry v-if="showNew"></NewEntry>    
  </section>
  <ErrorBar @reset-error="resetErrors" :usermessage="result" interval="2000"></ErrorBar>
</template>