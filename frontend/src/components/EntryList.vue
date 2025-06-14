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
  }
}
</script>

<template>
  <div id="div-monthly" class="work-entry">
    <h1>{{ headline }}</h1>
      <li v-for="item in reminders">
        {{ item.reminder.description }}, {{ item.next_occurrance }} 
        <button  @click="emitDelete(item.reminder.id)">Del</button>
        <button  @click="emitEdit({isnew: false, id: item.reminder.id})">Edit</button>
      </li>    
  </div>
</template>