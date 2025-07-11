<script>
import ErrorBar from './components/ErrorBar.vue';
import EntryList from './components/EntryList.vue'
import Navigation from './components/Navigation.vue'
import EditEntry from './components/EditEntry.vue';
import About from './components/About.vue';
import { monthSelected, newSelected, allSelected, aboutSelected } from './components/globals';
import { ReminderAPI, getDefaultReminder, Reminder } from './components/reminderapi';
import ConfirmationDialog from './components/ConfirmationDialog.vue';


export default {
  data() {
    return {
      overviewEntries: [],
      entriesInMonth: [],
      result: "",
      allRecipients: [],
      displayNameToId: {},
      idToDisplayName: {},
      currentComponent: monthSelected,
      monthToSearch: new Date().getMonth() + 1,
      yearToSearch: new Date().getFullYear(),
      apiURL: import.meta.env.VITE_API_URL,
      api: new ReminderAPI(import.meta.env.VITE_API_URL, ""),
      editData: "",
      reminderCount: 0,
      metrics: {},
      disableSave: false,
      defaultRecipientId: ""
    }
  },
  methods: {
    async deleteReminder(delNotification) {
      await this.deleteReminderAndSwitch(delNotification, this.currentComponent)
    },
    async deleteAndSwitchToNew(delNotification) {
      await this.deleteReminderAndSwitch(delNotification, newSelected)
    },
    async deleteReminderAndSwitch(delNotification, newPage) {
      const ok = await this.$refs.confirmationDialog.show('MobileNotifier', `Soll "${delNotification.description}" gelöscht werden?`, 'Löschen')

      if (ok) {
        let res = await this.api.deleteReminder(delNotification.id);
        if (res.error) {
          this.setErrorMessage("Eintrag konnte nicht gelöscht werden")
          return
        }

        await this.switchComponents(newPage)
      }
    },
    async saveReminder(saveData) {
      this.disableSave = true

      try {
        let h = saveData.reminderData
        let apiResult = null
        let savedId = ""

        if (saveData.id === null) {
          apiResult = await this.api.createNewReminder(h)
        } else {
          apiResult = await this.api.updateReminder(h, saveData.id)
        }

        if (apiResult.error) {
          this.setErrorMessage("Daten konten nicht gespeichert werden")
          return
        }

        savedId = apiResult.data
        this.editData = new Reminder(savedId, h.kind, h.param, h.warning_at, h.spec, h.description, h.recipients)
        this.setErrorMessage("Daten gespeichert")
      }
      finally
      {
        this.disableSave = false;
      }
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
        this.displayNameToId = {};
        this.idToDisplayName = {};
        return
      }

      let displayNames = []
      let toIdHelp = {};
      let toNameHelp = {};

      for (let i of res.data.all_recipients) {
        displayNames.push(i.display_name)
        toIdHelp[i.display_name] = i.id;
        toNameHelp[i.id] = i.display_name;
      }

      this.defaultRecipientId = res.data.default_id
      this.allRecipients = displayNames
      this.idToDisplayName = toNameHelp;
      this.displayNameToId = toIdHelp;
    },
    async getApiInfo() {
      let res = await this.api.getApiInfo()
      if (res.error) {
        this.setErrorMessage("Kann API info nicht abrufen");
        return;
      }

      this.apiVersion = res.data.version_info;
      this.apiTimeZone = res.data.time_zone;
      this.reminderCount = res.data.reminder_count;
      this.metrics = res.data.metrics;
    },
    resetErrors() {
      this.result = "";
    },
    setErrorMessage(msg) {
      this.result = msg;
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
    async incMonth() {
      if (this.monthToSearch == 12) {
        this.monthToSearch = 1;
        this.yearToSearch++;
      } else {
        this.monthToSearch++;
      }
      await this.getEventsInMonth();
    },
    async decMonth() {
      if (this.monthToSearch == 1) {
        this.monthToSearch = 12;
        this.yearToSearch--;
      } else {
        this.monthToSearch--;
      }
      await this.getEventsInMonth();
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
        this.editData = getDefaultReminder(this.defaultRecipientId);
      }

      await this.showComponents(value)
    },
    testCurrentComponent(refVal) {
      return this.currentComponent === refVal;
    },
    testAll() {
      return this.testCurrentComponent(allSelected);
    },
    testAbout() {
      return this.testCurrentComponent(aboutSelected);
    },
    testNew() {
      return this.testCurrentComponent(newSelected);
    },
    testMonth() {
      return this.testCurrentComponent(monthSelected);
    },
    async showComponents(value) {
      this.currentComponent = value
      this.result = "";

      if (value === monthSelected) {
        await this.getEventsInMonth()
      }

      if (value === allSelected) {
        await this.getOverview()
      }

      if (value === aboutSelected) {
        await this.getApiInfo()
      }
    }
  },
  components: {
    EntryList,
    Navigation,
    EditEntry,
    ErrorBar,
    About,
    ConfirmationDialog
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
    <table>
      <tr>
        <th width="12%"></th>
        <th></th>
      </tr>
      <tr>
        <td colspan="2">
          <Navigation @select-nav="switchComponents" :currentstate="currentComponent"></Navigation>
        </td>
      </tr>
      <tr>
        <td>
          Hinweise:
        </td>
        <td>
          <ErrorBar @reset-error="resetErrors" :usermessage="result" interval="2000"></ErrorBar>
        </td>
      </tr>
    </table>
  </section>

  <section class="work-items">
    <EntryList :reminders="overviewEntries" v-if="testAll()" headline="Alle Ereignisse" 
      @edit-id="editReminder" @delete-id="deleteReminder">
    </EntryList>
    <div v-if="testMonth()">
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
      <button id="nextmonth" @click="incMonth">Nächster Monat</button>
      <button id="prevmonth" @click="decMonth">Voriger Monat</button>
      <EntryList :reminders="entriesInMonth" headline="Ereignisse im gewählten Monat" 
        @edit-id="editReminder" @delete-id="deleteReminder">
      </EntryList>
    </div>
    <EditEntry v-if="testNew()" :allrecipients="allRecipients" :editdata="editData"
      :nametoid="displayNameToId" :idtoname="idToDisplayName" :disablesave="disableSave"
      @error-occurred="setErrorMessage" @delete-id="deleteAndSwitchToNew" @save-data="saveReminder">
    </EditEntry>
    <About v-if="testAbout()" :clienttz="apiTimeZone" :versioninfo="apiVersion" :apilink="apiURL"
      :elemcount="reminderCount" :metrics="metrics">
    </About>
    <ConfirmationDialog ref="confirmationDialog"></ConfirmationDialog>
  </section>
</template>