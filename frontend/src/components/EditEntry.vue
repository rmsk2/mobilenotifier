<script>
import { reminderAnniversary, ReminderData, reminderMonthly, reminderOneShot, reminderWeekly } from './reminderapi';
import { warningMorningBefore, warningNoonBefore, warningEveningBefore, warningWeekBefore, warningSameDay } from './reminderapi';
import { DeleteNotification, isLeapYear, incDay, decDay, sucMonth, predMonth, performDateCorrection } from './globals';


export default {
  data() {
    return {
      reminderOneShot: reminderOneShot,
      reminderAnniversary: reminderAnniversary,
      reminderMonthly: reminderMonthly,
      reminderWeekly: reminderWeekly,
      warningMorningBefore: warningMorningBefore,
      warningNoonBefore: warningNoonBefore,
      warningEveningBefore: warningEveningBefore,
      warningWeekBefore: warningWeekBefore,
      warningSameDay: warningSameDay,


      id: this.editdata.id,
      kind: this.editdata.kind,
      param: this.editdata.param,
      warningAt: this.editdata.warning_at,
      timeOfEvent: new Date(this.editdata.spec),
      description: this.editdata.description,
      recipients: this.editdata.recipients,
      month: new Date(this.editdata.spec).getMonth() + 1,
      year: new Date(this.editdata.spec).getFullYear(),
      hours: new Date(this.editdata.spec).getHours(),
      minutes: new Date(this.editdata.spec).getMinutes(),
      day: new Date(this.editdata.spec).getDate()
    }
  },
  watch: {
    editdata: {
        handler(newVal){
          this.copyData(newVal)
        },
        immediate: true
    },
    allrecipients: {
        handler(newVal){
          // make sure data is available as early as possible
          // simply defining the watcher as immediate allows 
          // this to happen
        },
        immediate: true
    },
    disablesave: {
        handler(newVal){
          // make sure data is available as early as possible
          // simply defining the watcher as immediate allows 
          // this to happen
        },
        immediate: true
    }
  },
  props: ['editdata', 'allrecipients', 'disablesave'],
  emits: ['error-occurred', 'delete-id', 'save-data'],
  computed: {
    weekDay() {
      let d = new Date(this.year, this.month-1, this.day);
      let days = ["Sonntag", "Montag", "Dienstag", "Mittwoch", "Donnerstag", "Freitag", "Samstag"]
      return days[d.getDay()]
    },
    recipientNames() {
      let res = []

      for (let i of this.allrecipients) {
        res.push(i.display_name)
      }

      return res
    },
    mapIdToName() {
      let res = {};

      for (let i of this.allrecipients) {
        res[i.id] = i.display_name
      }

      return res;
    },
    mapNameToId() {
      let res = {};

      for (let i of this.allrecipients) {
        res[i.display_name] = i.id
      }

      return res;
    }
  },
  methods: {
    nextDay() {
      let res = incDay(this.day, this.month, this.year)
      this.day = res.day
      this.month = res.month
      this.year = res.year
    },
    prevDay() {
      let res = decDay(this.day, this.month, this.year)
      this.day = res.day
      this.month = res.month
      this.year = res.year
    },
    nextMonth() {
      let res = sucMonth(this.day, this.month, this.year)
      this.day = res.day
      this.month = res.month
      this.year = res.year
    },
    prevMonth() {
      let res = predMonth(this.day, this.month, this.year)
      this.day = res.day
      this.month = res.month
      this.year = res.year
    },
    daySelected(event) {
      this.day = Number(event.target.value)
      this.performDateCorrection2()
    },
    monthSelected(event) {
      this.month = Number(event.target.value)
      this.performDateCorrection2()
    },
    yearSelected(event) {
      this.year = Number(event.target.value)
      this.performDateCorrection2()
    },
    performDateCorrection2() {
      let t = performDateCorrection(this.day, this.month, this.year)
      this.day = t
    },
    makeNumeric() {
      let h = []
      for (let i in this.warningAt) {
        h.push(Number(this.warningAt[i]))
      }
      this.warningAt = h
      this.kind = Number(this.kind)
    },
    async deleteEntry() {
      this.$emit('delete-id',  new DeleteNotification(this.id, this.description))
    },
    createNew() {
      return this.id === null
    },
    namesToIds(recipientNames) {
      let res = [];

      for (let i of recipientNames) {
        res.push(this.mapNameToId[i]);
      }

      return res
    },
    idsToNames(recipientIds) {
      let res = [];

      for (let i of recipientIds) {
        res.push(this.mapIdToName[i]);
      }

      return res;
    },
    validate() {
      if (this.warningAt.length == 0) {
        return {ok: false, msg: "Keine Warnung eingetragen"}
      }
      
      if (this.recipients.length == 0) {
        return {ok: false, msg: "Keine Empfänger angegeben"}
      }

      if (this.param > 23) {
        return {ok: false, msg: "Der Vorlauf kann maximal 23 Stunden sein"}
      }

      if (this.description.length > 140) {
        return {ok: false, msg: "Die Beschreibung passt in keine SMS"}
      }

      if ((this.month == 2) && (this.day > 29)) {
        return {ok: false, msg: `Es gibt keinen ${this.day}.ten Februar`}
      }

      if ((this.month == 2) && !isLeapYear(this.year) && (this.day > 28)) {
        return {ok: false, msg: `${this.year} ist kein Schaltjahr`}
      }

      let shortMonths = new Set(["4", "6", "9", "11"]);
      if (shortMonths.has(this.month.toString()) && (this.day > 30)) {
        return {ok: false, msg: `Es gibt keinen ${this.day}.ten in diesem Monat`}
      }

      return {ok: true, msg: ""}
    },
    async saveData() {
      let valRes = this.validate();
      if (!valRes.ok) {
        this.$emit('error-occurred', valRes.msg);
        return;
      }

      this.makeNumeric()
      let h = new Date(this.year, this.month-1, this.day, this.hours, this.minutes, 0)
      let utcDate = new Date(h.toISOString())      
      let remData = new ReminderData(this.kind, this.param, this.warningAt, utcDate, this.description, this.namesToIds(this.recipients))

      this.$emit('save-data', {id: this.id, reminderData: remData})
    },
    copyData(from) {
      let d = new Date(from.spec);
      this.id = from.id;
      this.kind = from.kind;
      this.param = from.param;
      this.warningAt = from.warning_at;
      this.timeOfEvent = d;
      this.description = from.description;
      this.recipients = this.idsToNames(from.recipients);
      this.day = d.getDate();
      this.month = d.getMonth() + 1;
      this.year = d.getFullYear();
      this.hours = d.getHours();
      this.minutes = d.getMinutes();
    }
  }
}
</script>

<template>
  <div id="div-edit-entry" class="work-entry">
    <h2 v-if="createNew()">Neues Ereignis erstellen</h2>
    <h2 v-else>Bestehendes Ereignis bearbeiten</h2>

    <fieldset>
      <legend>
        <b>Basisdaten des Ereignisses</b>
      </legend>
      <table class="edit-entry-table">
        <tr>
          <td>Shortcuts</td>
          <td>              
              <button @click="nextDay">Einen Tag vor</button>
              <button @click="prevDay">Einen Tag zurück</button>
              <button @click="nextMonth">Einen Monat vor</button>
              <button @click="prevMonth">Einen Monat zurück</button>
          </td>
        </tr>
        <tr>
          <td>Zeitpunkt</td>
          <td>
            <div id="eventtime" name="eventtime">              
              <select name="dayselect" :value="day"  @change="daySelected" id="dayselect">
                <option value="1">01</option>
                <option value="2">02</option>
                <option value="3">03</option>
                <option value="4">04</option>
                <option value="5">05</option>
                <option value="6">06</option>
                <option value="7">07</option>
                <option value="8">08</option>
                <option value="9">09</option>
                <option value="10">10</option>
                <option value="11">11</option>
                <option value="12">12</option>
                <option value="13">13</option>
                <option value="14">14</option>
                <option value="15">15</option>
                <option value="16">16</option>
                <option value="17">17</option>
                <option value="18">18</option>
                <option value="19">19</option>
                <option value="20">20</option>
                <option value="21">21</option>
                <option value="22">22</option>
                <option value="23">23</option>
                <option value="24">24</option>
                <option value="25">25</option>
                <option value="26">26</option>
                <option value="27">27</option>
                <option value="28">28</option>
                <option value="29">29</option>
                <option value="30">30</option>
                <option value="31">31</option>
              </select>  

              <select name="monthselect" :value="month" @change="monthSelected" id="monthselect">
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

              <input type="number" size="6" id="yearin" name="yearin" :value="year" @change="yearSelected"></input>

              <select name="hourselect" v-model="hours" id="hourselect">
                <option value="0">00</option>
                <option value="1">01</option>
                <option value="2">02</option>
                <option value="3">03</option>
                <option value="4">04</option>
                <option value="5">05</option>
                <option value="6">06</option>
                <option value="7">07</option>
                <option value="8">08</option>
                <option value="9">09</option>
                <option value="10">10</option>
                <option value="11">11</option>
                <option value="12">12</option>
                <option value="13">13</option>
                <option value="14">14</option>
                <option value="15">15</option>
                <option value="16">16</option>
                <option value="17">17</option>
                <option value="18">18</option>
                <option value="19">19</option>
                <option value="20">20</option>
                <option value="21">21</option>
                <option value="22">22</option>
                <option value="23">23</option>
              </select>

              <select name="minuteselect" v-model="minutes" id="minuteselect">
                <option value="0">00</option>
                <option value="1">01</option>
                <option value="2">02</option>
                <option value="3">03</option>
                <option value="4">04</option>
                <option value="5">05</option>
                <option value="6">06</option>
                <option value="7">07</option>
                <option value="8">08</option>
                <option value="9">09</option>
                <option value="10">10</option>
                <option value="11">11</option>
                <option value="12">12</option>
                <option value="13">13</option>
                <option value="14">14</option>
                <option value="15">15</option>
                <option value="16">16</option>
                <option value="17">17</option>
                <option value="18">18</option>
                <option value="19">19</option>
                <option value="20">20</option>
                <option value="21">21</option>
                <option value="22">22</option>
                <option value="23">23</option>
                <option value="24">24</option>
                <option value="25">25</option>
                <option value="26">26</option>
                <option value="27">27</option>
                <option value="28">28</option>
                <option value="29">29</option>
                <option value="30">30</option>
                <option value="31">31</option>
                <option value="32">32</option>
                <option value="33">33</option>
                <option value="34">34</option>
                <option value="35">35</option>
                <option value="36">36</option>
                <option value="37">37</option>
                <option value="38">38</option>
                <option value="39">39</option>
                <option value="40">40</option>
                <option value="41">41</option>
                <option value="42">42</option>
                <option value="43">43</option>
                <option value="44">44</option>
                <option value="45">45</option>
                <option value="46">46</option>
                <option value="47">47</option>
                <option value="48">48</option>
                <option value="49">49</option>
                <option value="50">50</option>
                <option value="51">51</option>
                <option value="52">52</option>
                <option value="53">53</option>
                <option value="54">54</option>
                <option value="55">55</option>
                <option value="56">56</option>
                <option value="57">57</option>
                <option value="58">58</option>
                <option value="59">59</option>
              </select>
              {{ weekDay }}
            </div>
          </td>
        </tr>
        <tr>
          <td>Beschreibung</td>
          <td>
            <input type="text" id="desc" name="desc" size="80" v-model="description"></input><br>
          </td>
        </tr>
        <tr>
          <td>Art des Ereignisses</td>
          <td>
            <select name="kindselect" v-model="kind" id="kindselect">
              <option :value="reminderOneShot">Einmaliges Ereignis</option>
              <option :value="reminderAnniversary">Jährlich wiederkehrendes Ereignis</option>
              <option :value="reminderMonthly">Monatlich wiederkehrendes Ereignis</option>
              <option :value="reminderWeekly">Wöchentlich wiederkehrendes Ereignis</option>
            </select>
          </td>
        </tr>
      </table>
    </fieldset>

    <fieldset id="warningtypes" name="warningtypes">
      <legend>
        <b>Gewünschte Vorwarnung</b>
      </legend>
      <table style="border-collapse: collapse;">
        <tr>
          <td>
            <input type="checkbox" v-model="warningAt" name="warningAt" :value="warningMorningBefore" />Am Morgen des vorigen Tages<br/>
          </td>
          <td>
            <input type="checkbox" v-model="warningAt" name="warningAt" :value="warningNoonBefore" />Am Mittag des vorigen Tages<br/>
          </td>
        </tr>
        <tr>
          <td>
            <input type="checkbox" v-model="warningAt" name="warningAt" :value="warningEveningBefore" />Am Abend des vorigen Tages<br/>
          </td>
          <td>
            <input type="checkbox" v-model="warningAt" name="warningAt" :value="warningWeekBefore" />Eine Woche vor dem Ereignis<br/>
          </td>
        </tr>
        <tr>
          <td>
            <input type="checkbox" v-model="warningAt" name="warningAt" :value="warningSameDay" />Am Tag des Ereignisses<br/>
          </td>
          <td>
            <label for="param">Vorlauf in Stunden bei Warnung am gleichen Tag:</label>
            <input type="number" size="3" id="param" name="param" v-model="param"></input>
          </td>
        </tr>
      </table>
    </fieldset>

    <fieldset>
      <legend>
        <b>Wer soll gewarnt werden</b>
      </legend>
      <div v-for="r in recipientNames">
        <input  type="checkbox" v-model="recipients" name="selrecipients" :value="r">{{ r }}</input>
      </div>
    </fieldset>

    <button @click="saveData" :disabled="disablesave">Daten speichern</button><button v-if="!createNew()" @click="deleteEntry" :disabled="disablesave">Ereignis löschen</button>
  </div>
</template>