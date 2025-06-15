<script>
import ErrorBar from './components/ErrorBar.vue';
import EntryList from './components/EntryList.vue'
import Navigation from './components/Navigation.vue'
import EditEntry from './components/EditEntry.vue';
import { monthSelected, newSelected, allSelected } from './components/globals';
import { ReminderAPI, getDefaultReminder } from './components/reminderapi';


export default {
  data() {
    return {
      showMonthly: true,
      showAll: false,
      showNew: false,
      overviewEntries: [],
      entriesInMonth: [],
      result: "",
      allRecipients: [],
      currentComponent: monthSelected,
      monthToSearch: new Date().getMonth() + 1,
      yearToSearch: new Date().getFullYear(),
      api: new ReminderAPI(import.meta.env.VITE_API_URL, ""),
      editData: "kacke"
    }
  },
  methods: {
    async deleteReminder(id) {
      let res = await this.api.deleteReminder(id);
      if (res.error) {
        this.setErrorMessage("Eintrag konnte nicht gelöscht werden")
        return
      }

      await this.redraw()
    },
    async redraw() {
      await this.showComponents(this.currentComponent)
    },
    async editReminder(info) {
      let res = await this.api.readReminder(info.id)
      if (res.error) {
        this.setErrorMessage("Kann Ereignis nicht laden")
        return
      }

      if (!res.data.found) {
        this.setErrorMessage("Ereignis-ID nicht vorhanden")
        return
      }

      this.editData = res.data.data
      this.showComponents(newSelected)
    },
    async getRecipients() {
      let res = await this.api.getRecipients()
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
      let res = await this.api.getOverview();      
      if (res.error) {
        this.setErrorMessage("Kann Übersicht nicht abrufen");
        this.overviewEntries = [];
        return;
      }

      this.overviewEntries = res.data;
    },
    async getEventsInMonth() {
      let res = await this.api.getEventsInMonth(this.monthToSearch, this.yearToSearch);
      
      if (res.error) {
        this.setErrorMessage("Kann Ereignisse nicht abrufen");
        this.entriesInMonth = [];
        return
      }
      
      this.entriesInMonth = res.data;      
    },
    async switchComponents(value) {
      if (value == newSelected) {        
        this.editData = getDefaultReminder(this.allRecipients[0]);
      }      

      await this.showComponents(value)
    },
    async showComponents(value) {
      this.currentComponent = value
      this.makeAllComponentsInvisible();
      this.result = "";

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
    EntryList,
    Navigation,
    EditEntry,
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
    <Navigation @select-nav="switchComponents"></Navigation>
  </section>

  <p/>
  <section class="work-items">
    <EntryList :reminders="overviewEntries" v-if="showAll" headline="Alle Ereignisse"
      @edit-id="editReminder" 
      @delete-id="deleteReminder">
    </EntryList>
    <div v-if="showMonthly">
      <select name="months" v-model="monthToSearch" @change="redraw" id="selectmonth">
        <option value="1">Januar</option>
        <option value="2">Februar</option>
        <option value="3">März</option>
        <option value="4">April</option>
        <option value="5">Mai</option>
        <option value="6">Juni</option>
        <option value="7">Juli</option>
        <option value="8">August</option>
        <option value="9">September</option>
        <option value="10">Oktober</option>
        <option value="11">November</option>
        <option value="12">Dezember</option>
      </select>  
      <input type="number" v-model="yearToSearch" @change="redraw" name="yearentry" id="yearentry">
      <EntryList :reminders="entriesInMonth"  headline="Ereignisse im gewählten Monat"
        @edit-id="editReminder"
        @delete-id="deleteReminder">
      </EntryList>
    </div>
    <EditEntry v-if="showNew"
      :api="api" :allrecipients="allRecipients" :editdata="editData"
      @error-occurred="setErrorMessage">
    </EditEntry>
  </section>
  <ErrorBar @reset-error="resetErrors" :usermessage="result" interval="2000"></ErrorBar>
</template>