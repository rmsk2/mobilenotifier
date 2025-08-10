<script>
import { addrClassIfttt, addrClassMail } from './reminderapi';

export default {
  data() {
    return {
    }
  },
  props: ['allrecipients'],
  methods: {
    recipientsAvailable() {
      return this.allrecipients.length !== 0;
    },    
  },
}
</script>

<template>
  <div id="div-entry-list" class="work-entry">
    <h1>Liste der Empfänger</h1>
    <table id="table-found-events" class="table-list-events" v-if="recipientsAvailable()">
      <tr class="table-list-events-row">
        <th class="table-list-events-header">Name</th>
        <th class="table-list-events-header">Adresse</th>
        <th class="table-list-events-header">Typ</th>
        <th class="table-list-events-header">Default</th>
      </tr>
      <tr class="table-list-events-row" v-for="item in allrecipients">
        <td class="table-list-events-elem" :class="item.cls"><span class="list-text">{{ item.display_name }}</span></td> 
        <td class="table-list-events-elem" :class="item.cls"><span class="list-text">{{ item.address }}</span></td> 
        <td class="table-list-events-elem" :class="item.cls"><span class="list-text">{{ item.addr_type }}</span></td> 
        <td class="table-list-events-elem" :class="item.cls"><span class="list-text">{{ item.is_default }}</span></td> 
      </tr>
    </table>
    <div id="no-found-events" class="text-no-events" v-if="!recipientsAvailable()">
      Keine Empfänger gefunden
    </div>
  </div>
</template>