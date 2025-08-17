<script>
import { addrClassIfttt, addrClassMail, Recipient } from './reminderapi';

export default {
  data() {
    return {
      addrTypeIfttt: addrClassIfttt,
      addrTypeMail: addrClassMail,

      editId: null,
      displayName: "",
      isDefault: false,
      addressType: addrClassIfttt,
      address: "",
      editButtonText: "Empfängerliste ändern"
    }
  },
  props: ['allrecipients', 'editvisible'],
  emits: ['upsert-entry', 'delete-id', 'error-occurred', 'toggle-edit'],
  methods: {
    recipientsAvailable() {
      return this.allrecipients.length !== 0;
    },
    upsertRecipient(entryData) {
      let checkRes = this.validate()
      if (!checkRes.ok) {
        this.$emit('error-occurred', checkRes.msg);
        return;
      }

      if (this.isDefault === "true") {
        this.isDefault = true
      }
      if (this.isDefault === "false") {
        this.isDefault = false
      }
      entryData = new Recipient(this.addressType, this.address, this.displayName, this.editId, this.isDefault)
      this.$emit('upsert-entry', entryData);
    },
    procNewEntry() {
      this.editId = null
      this.address = ""
      this.addressType = this.addrTypeIfttt
      this.isDefault = false
      this.displayName = ""
    },
    procDelete(id) {
      this.$emit('delete-id', id);
    },
    procEdit(id) {
      this.editId = id
      this.address = this.recipientDict[id].address
      this.addressType = this.recipientDict[id].addr_type
      this.isDefault = this.recipientDict[id].is_default
      this.displayName = this.recipientDict[id].display_name
    },
    toggleEditable() {
      this.$emit("toggle-edit")
    },
    setButtonText(val) {
      if (val) {
        this.editButtonText = "Empfängerliste nur betrachten"
      } else {
        this.editButtonText = "Empfängerliste ändern"
      }
    },
    getDefaultText(val) {
      if (val) {
        return "Ja"
      } else {
        return "Nein"
      }
    },
    validate() {
      if ((this.address === "") || (this.displayName === "")) {
        return ({"ok": false, "msg":"Adresse und Anzeigename müssen gefüllt sein"})
      }

      if (this.allrecipients.filter(i => i.id !== this.editId).map(i => i.display_name).some(i => i === this.displayName)) {
        return ({"ok": false, "msg":"Der Anzeigename muss eindeutig sein"})
      }

      return {"ok": true, "msg":""};
    }
  },
  watch: {
    allrecipients: {
        handler(newVal){
          // make sure data is available as early as possible
          // simply defining the watcher as immediate allows
          // this to happen
          this.procNewEntry()
        },
        immediate: true
    },
    editvisible: {
        handler(newVal){
          // make sure data is available as early as possible
          // simply defining the watcher as immediate allows
          // this to happen
          this.setButtonText(newVal)
        },
        immediate: true
    }
  },
  computed: {
    recipientDict() {
      let res = {}

      for (let e of this.allrecipients) {
        res[e.id] = e
      }

      return res;
    }
  }
}
</script>

<template>
  <div id="div-entry-list" class="work-entry">
    <h1>Liste der Empfänger</h1>
    <button @click="toggleEditable()">{{ editButtonText }}</button>
    <p/>
    <table id="table-found-events" class="table-list-events" v-if="recipientsAvailable()">
      <tr class="table-list-events-row">
        <th class="table-list-events-header">Name</th>
        <th class="table-list-events-header">Adresse</th>
        <th class="table-list-events-header">Typ</th>
        <th class="table-list-events-header">Default</th>
        <th v-if="editvisible" class="table-list-events-header">Bearbeiten</th>
      </tr>
      <tr class="table-list-events-row" v-for="item in allrecipients">
        <td class="table-list-events-elem" :class="item.cls"><span class="list-text">{{ item.display_name }}</span></td> 
        <td class="table-list-events-elem" :class="item.cls"><span class="list-text">{{ item.address }}</span></td> 
        <td class="table-list-events-elem" :class="item.cls"><span class="list-text">{{ item.addr_type }}</span></td> 
        <td class="table-list-events-elem" :class="item.cls"><span class="list-text">{{ getDefaultText(item.is_default) }}</span></td>
        <td v-if="editvisible" class="table-list-events-elem table-list-buttons">
          <button  @click="procDelete(item.id)">Del</button>
          <button  @click="procEdit(item.id)">Edit</button>
        </td>
      </tr>
    </table>
    <div id="no-found-events" class="text-no-events" v-if="!recipientsAvailable()">
      Keine Empfänger gefunden
    </div>
  </div>
  <p/>
  <div v-if="editvisible">
    <table id="table-found-events" class="table-list-events">
      <tr class="table-list-events-row">
        <th class="table-list-events-header">Neuer Name</th>
        <th class="table-list-events-header">neue Adresse</th>
        <th class="table-list-events-header">Neuer Typ</th>
        <th class="table-list-events-header">Neuer Default</th>
      </tr>
      <tr class="table-list-events-row">
        <td class="table-list-events-elem"><input type="text" id="desc" name="desc" size="20" class="list-text" v-model="displayName"></input></td>
        <td class="table-list-events-elem"><input type="text" id="desc" name="desc" size="30" class="list-text" v-model="address"></input></td>
        <td class="table-list-events-elem table-list-buttons">
          <select name="typeselect" v-model="addressType" id="typeselect">
            <option class="list-text" :value="addrTypeIfttt">IFTTT</option>
            <option class="list-text" :value="addrTypeMail">Mail</option>
          </select>
        </td>
        <td class="table-list-events-elem table-list-buttons">
          <select name="defaultselect" v-model="isDefault" id="defaultselect">
            <option class="list-text" value="true">Ja</option>
            <option class="list-text" value="false">Nein</option>
          </select>
        </td>
      </tr>
    </table>
    <p/>
    <button @click="procNewEntry">Alle Eingabewerte zurücksetzen</button><button @click="upsertRecipient">
      <span v-if="this.editId===null">Neuen Eintrag erstellen</span>
      <span v-if="this.editId!==null">Eintrag aktualisieren</span>
    </button>
  </div>
</template>