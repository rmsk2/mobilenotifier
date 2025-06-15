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
        let dt = d.toLocaleDateString("de-DE", options);
        let tt = d.toLocaleTimeString();
        res.push({id: e.reminder.id, text: `${dt} um ${tt} Uhr - ${t}`});
      }
      
      return res;
    }
  }
}
</script>

<template>
  <div id="div-entry-list" class="work-entry">
    <h1>{{ headline }}</h1>
      <li v-for="item in formattedEvents">
        {{ item.text }} 
        <button  @click="emitDelete(item.id)">Del</button>
        <button  @click="emitEdit({isnew: false, id: item.id})">Edit</button>
      </li>    
  </div>
</template>