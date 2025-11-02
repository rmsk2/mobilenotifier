<script>
import { reminderAnniversary, reminderMonthly, reminderWeekly } from './reminderapi';
import { DeleteNotification } from './globals';

export default {
  data() {
    return {
    }
  },
  props: ['reminders', 'headline', 'clienttime'],
  emits: ['edit-id', 'delete-id'],
  methods: {
    emitEdit(info) {
      this.$emit('edit-id', info);
    },
    emitDelete(id) {
      this.$emit('delete-id', id);
    },
    eventsAvailable() {
      return this.reminders.length !== 0;
    },
    makeNotification(id, description) {
      return new DeleteNotification(id, description)
    }
  },
  computed: {
    formattedEvents() {
      let options = { weekday: 'long', year: 'numeric', month: 'long', day: 'numeric' };
      let res = []

      for (let e of this.reminders) {
        let t = e.reminder.description;
        let d = new Date(e.next_occurrance);
        let td = d.toLocaleDateString("de-DE", options);
        let tt = d.toLocaleTimeString().substring(0, 5);
        let cl = "list-not-anniversary";

        if (e.reminder.kind === reminderAnniversary) {
          cl = "list-anniversary";
        }

        if (e.reminder.kind === reminderMonthly) {
          cl = "list-monthly";
        }

        if (e.reminder.kind === reminderWeekly) {
          cl = "list-weekly";
        }        

        res.push({id: e.reminder.id, textDate: td, textTime: tt, text: t, cls: cl});
      }
      
      return res;
    },
    currentDate() {
      let options = { weekday: 'long', year: 'numeric', month: 'long', day: 'numeric' };
      let d = new Date(this.clienttime);
      return d.toLocaleDateString("de-DE", options);
    }
  }
}
</script>

<template>
  <div id="div-entry-list" class="work-entry">
    <h1>{{ headline }}</h1>
    <span class="text-no-events">Heute ist {{ currentDate }}</span>
    <p/>
    <table id="table-found-events" class="table-list-events" v-if="eventsAvailable()">
      <tr class="table-list-events-row">
        <th class="table-list-events-header">Am</th>
        <th class="table-list-events-header">Um</th>
        <th class="table-list-events-header">Ereignis</th>
        <th class="table-list-events-header">Bearbeiten</th>
      </tr>
      <tr class="table-list-events-row" v-for="item in formattedEvents">
        <td class="table-list-events-elem" :class="item.cls"><span class="list-text">{{ item.textDate }}</span></td> 
        <td class="table-list-events-elem" :class="item.cls"><span class="list-text">{{ item.textTime }}</span></td> 
        <td class="table-list-events-elem" :class="item.cls"><span class="list-text">{{ item.text }}</span></td> 
        <td class="table-list-events-elem table-list-buttons">
          <button  @click="emitDelete(makeNotification(item.id, item.text))">Del</button>
          <button  @click="emitEdit({isnew: false, id: item.id})">Edit</button>
        </td>
      </tr>
    </table>
    <div id="no-found-events" class="text-no-events" v-if="!eventsAvailable()">
      Keine Ereignisse gefunden
    </div>
  </div>
</template>