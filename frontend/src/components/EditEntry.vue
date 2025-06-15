<script>
import { reminderAnniversary, ReminderData, reminderOneShot } from './reminderapi';
import { warningMorningBefore, warningNoonBefore, warningEveningBefore, warningWeekBefore, warningSameDay } from './reminderapi';
import { getDefaultReminder } from './reminderapi';


export default {
  data() {
    return {
      reminderOneShot: reminderOneShot,
      reminderAnniversary: reminderAnniversary,
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
      seconds: new Date(this.editdata.spec).getSeconds(),
      day: new Date(this.editdata.spec).getDate(),
    }
  },
  watch: {
    editdata: {
        handler(newVal){
          this.copyData(newVal)
        },
        immediate: true
    }
  },  
  props: ['editdata', 'api', 'allrecipients'],
  emits: ['error-occurred'],
  methods: {
    makeNumeric() {
      let h = []
      for (let i in this.warningAt) {
        h.push(Number(this.warningAt[i]))
      }
      this.warningAt = h
      this.kind = Number(this.kind)
    },
    async deleteEntry() {
      let res = this.api.deleteReminder(this.id)
      if (res.error) {  
        this.$emit('error-occurred', "Daten konten nicht gelöscht werden");
        return;
      } 
      
      this.$emit('error-occurred', "Daten gelöscht")
      this.copyData(getDefaultReminder(this.allrecipients[0]))
    },
    createNew() {
      return this.id === null
    },
    validate() {
      if (this.warningAt.length == 0) {
        return {ok: false, msg: "Keine Warnung eingetragen"}
      }
      
      if (this.recipients.length == 0) {
        return {ok: false, msg: "Keine Empfänger angegeben"}
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

      let h = new Date(this.year, this.month-1, this.day, this.hours, this.minutes, this.seconds)
      let utcDate = new Date(h.toISOString())
      
      let remData = new ReminderData(this.kind, this.param, this.warningAt, utcDate, this.description, this.recipients)
      let res = null

      if (this.createNew()) {
        res = await this.api.createNewReminder(remData)
      } else {
        res = await this.api.updateReminder(remData, this.id)
      }

      if (res.error) {  
        this.$emit('error-occurred', "Daten konten nicht gespeichert werden")
        return
      } 

      this.$emit('error-occurred', "Daten gespeichert")
      this.id = res.data.id;
    },
    copyData(from) {
      let d = new Date(from.spec);
      this.id = from.id;
      this.kind = from.kind;
      this.param = from.param;
      this.warningAt = from.warning_at;
      this.timeOfEvent = d;
      this.description = from.description;
      this.recipients = from.recipients;
      this.day = d.getDate();
      this.month = d.getMonth() + 1;
      this.year = d.getFullYear();
      this.hours = d.getHours();
      this.minutes = d.getMinutes();
      this.seconds = d.getSeconds();
    }
  }
}
</script>

<template>
  <div id="div-edit-entry" class="work-entry">
    <h2 v-if="createNew()">Neues Ereignis erstellen</h2>
    <h2 v-else>Bestehendes Ereignis bearbeiten</h2>

    <label for="eventtime">Wann findet das Ereignis statt:</label>
    <div id="eventtime" name="eventtime">
      
      <select name="dayselect" v-model="day" id="dayselect">
        <option value="1">1</option>
        <option value="2">2</option>
        <option value="3">3</option>
        <option value="4">4</option>
        <option value="5">5</option>
        <option value="6">6</option>
        <option value="7">7</option>
        <option value="8">8</option>
        <option value="9">9</option>
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

      <select name="monthselect" v-model="month" id="monthselect">
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

      <input type="number" id="yearin" name="yearin" v-model="year"></input>

      <select name="hourselect" v-model="hours" id="hourselect">
        <option value="0">0</option>
        <option value="1">1</option>
        <option value="2">2</option>
        <option value="3">3</option>
        <option value="4">4</option>
        <option value="5">5</option>
        <option value="6">6</option>
        <option value="7">7</option>
        <option value="8">8</option>
        <option value="9">9</option>
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
        <option value="0">0</option>
        <option value="1">1</option>
        <option value="2">2</option>
        <option value="3">3</option>
        <option value="4">4</option>
        <option value="5">5</option>
        <option value="6">6</option>
        <option value="7">7</option>
        <option value="8">8</option>
        <option value="9">9</option>
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

      <select name="secondsselect" v-model="seconds" id="secondsselect">
        <option value="0">0</option>
        <option value="1">1</option>
        <option value="2">2</option>
        <option value="3">3</option>
        <option value="4">4</option>
        <option value="5">5</option>
        <option value="6">6</option>
        <option value="7">7</option>
        <option value="8">8</option>
        <option value="9">9</option>
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
    </div>
    <p/>

    <label for="desc">Beschreibung:</label>
    <input type="text" id="desc" name="desc" size="80" v-model="description"></input>
    <p/>

    <label for="kindselect">Art des Ereignisses:</label>
    <select name="kindselect" v-model="kind" id="kindselect">
      <option :value="reminderOneShot">Einmaliges Ereignis</option>
      <option :value="reminderAnniversary">Jährlich wiederkehrendes Ereignis</option>
    </select> 
    <p/>

    <label for="warningtypes">Gewünschte Vorwarnung:</label>
    <div id="warningtypes" name="warningtypes">
      <input type="checkbox" v-model="warningAt" name="warningAt" :value="warningMorningBefore" />Am Morgen des vorigen Tages<br/>
      <input type="checkbox" v-model="warningAt" name="warningAt" :value="warningNoonBefore" />Am Mittag des vorigen Tages<br/>
      <input type="checkbox" v-model="warningAt" name="warningAt" :value="warningEveningBefore" />Am Abend des vorigen Tages<br/>
      <input type="checkbox" v-model="warningAt" name="warningAt" :value="warningWeekBefore" />Eine Woche vor dem Ereignis<br/>
      <input type="checkbox" v-model="warningAt" name="warningAt" :value="warningSameDay" />Am Tag des Ereignisses<br/>
    </div>
    <p/>

    <label for="recipientnames">Wer soll gewarnt werden:</label>
    <div id="recipientnames" name="recipientnames">
        <li v-for="r in allrecipients">
          <input type="checkbox" v-model="recipients" name="selrecipients" :value="r">{{ r }}
        </li>
    </div>
    <p/>

    <label for="param">Vorlauf in Stunden bei Warnung am gleichen Tag:</label>
    <input type="number" id="param" name="param" v-model="param"></input>
    <p/>
    <button @click="saveData">Daten speichern</button><button v-if="!createNew()" @click="deleteEntry">Eintrag löschen</button>
  </div>
</template>