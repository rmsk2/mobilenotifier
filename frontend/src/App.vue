<script>
import ErrorBar from './components/ErrorBar.vue';
import EntryList from './components/EntryList.vue'
import Navigation from './components/Navigation.vue'
import EditEntry from './components/EditEntry.vue';
import About from './components/About.vue';
import { monthSelected, newSelected, allSelected, aboutSelected } from './components/globals';
import { ReminderAPI, getDefaultReminder } from './components/reminderapi';


export default {
  data() {
    return {
      showMonthly: true,
      showAll: false,
      showNew: false,
      showAbout: false,
      overviewEntries: [],
      entriesInMonth: [],
      result: "",
      allRecipients: [],
      currentComponent: monthSelected,
      monthToSearch: new Date().getMonth() + 1,
      yearToSearch: new Date().getFullYear(),
      api: new ReminderAPI(import.meta.env.VITE_API_URL, ""),
      editData: "kacke",
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
    async getApiInfo() {
      let res = await this.api.getApiInfo()
      if (res.error) {
        this.setErrorMessage("Kann API info nicht abrufen")
        return
      }

      this.apiVersion = res.data.version_info;
      this.apiTimeZone = res.data.time_zone;
    },
    resetErrors() {
      this.result = ""
    },
    setErrorMessage(msg) {
      this.result = msg
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
      this.result = "";

      if (value === monthSelected) {
        await this.getEventsInMonth()
        this.showMonthly = true;
        this.showAll = false;
        this.showNew = false;
        this.showAbout = false;
      }

      if (value === allSelected) {
        await this.getOverview()        
        this.showAll = true;
        this.showMonthly = false;
        this.showNew = false;
        this.showAbout = false;
      }
      
      if (value === newSelected) {
        this.showNew = true;
        this.showMonthly = false;
        this.showAll = false;
        this.showAbout = false;
      }

      if (value === aboutSelected) {
        this.showNew = false;
        this.showMonthly = false;
        this.showAll = false;
        this.showAbout = true;
      }
    }
  },
  components: {
    EntryList,
    Navigation,
    EditEntry,
    ErrorBar,
    About
  },
  async beforeMount() {
    await this.getApiInfo();
    await this.getRecipients();
    await this.showComponents(monthSelected);
  },  
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
    <About v-if="showAbout" 
      :clienttz="apiTimeZone" :versioninfo="apiVersion">
    </About>
  </section>
  <ErrorBar @reset-error="resetErrors" :usermessage="result" interval="2000"></ErrorBar>
</template>