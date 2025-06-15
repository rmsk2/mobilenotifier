<script>

export default {
  data() {
    return {
    }
  },
  props: ['reminders', 'headline'],
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
    }
  },
  computed: {
    formattedEvents() {
      let options = { weekday: 'long', year: 'numeric', month: 'long', day: 'numeric' };
      let res = []      

      console.log(this.reminders);

      for (let i in this.reminders) {
        let e = this.reminders[i];

        let t = e.reminder.description;
        let d = new Date(e.next_occurrance);
        let td = d.toLocaleDateString("de-DE", options);
        let tt = d.toLocaleTimeString();
        res.push({id: e.reminder.id, textDate: td, textTime: tt, text: t});
      }
      
      return res;
    }
  }
}
</script>

<template>
  <div id="div-entry-list" class="work-entry">
    <h1>{{ headline }}</h1>
    <table id="table-found-events" class="list-events" v-if="eventsAvailable()">
      <tr>
        <th>Am</th>
        <th>Um</th>
        <th>Ereignis</th>
        <th>Bearbeiten</th>
      </tr>
      <tr v-for="item in formattedEvents">
        <td>{{ item.textDate }}</td> 
        <td>{{ item.textTime }}</td> 
        <td>{{ item.text }}</td> 
        <td>
          <button  @click="emitDelete(item.id)">Del</button>
          <button  @click="emitEdit({isnew: false, id: item.id})">Edit</button>
        </td>
      </tr>
    </table>
    <div id="no-found-events" class="text-no-events" v-if="!eventsAvailable()">
      Keine Ereignisse gefunden
    </div>
  </div>
</template>