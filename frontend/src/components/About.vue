<script>
import { versionString } from './globals';

export default {
  data() {
    return {
      webUIVersion: versionString
    }
  },
  props: ['versioninfo', 'clienttz', 'apilink'],
  computed: {
    swaggerUrl() {
      return `${this.apilink}swagger/index.html`;
    }
  }
}
</script>

<template>
  <div id="div-about" class="about-component">
    <h2>Über diese Anwendung</h2>
    <div class="about-text">
    Mobile-Notifier (a.k.a das Gschmarri-Projekt). Eine Webanwendung zur Verwaltung von SMS-Erinnerungen. Geschrieben von Martin Grap in 2025.
    </div>
    <p/>
    <table class="about-table">
      <tr>
        <th class="table-about-header"><span class="about-text">Info</span></th>
        <th class="table-about-header"><span class="about-text">Wert</span></th>
      </tr>
      <tr>
        <td class="table-about-elem"><span class="about-text">WebUI-Version</span></td>
        <td class="table-about-elem"><span class="about-text">{{ webUIVersion }}</span></td>
      </tr>
      <tr>
        <td class="table-about-elem"><span class="about-text">Swagger-URL</span></td>
        <td class="table-about-elem"><span class="about-text"><a :href="swaggerUrl">{{swaggerUrl}}</a></span></td>
      </tr>
      <tr>
        <td class="table-about-elem"><span class="about-text">API-Version</span></td>
        <td class="table-about-elem"><span class="about-text">{{ versioninfo }}</span></td>
      </tr>
      <tr>
        <td class="table-about-elem"><span class="about-text">Zeitzone des Clients im Backend</span></td>
        <td class="table-about-elem"><span class="about-text">{{ clienttz }}</span></td>
      </tr>
    </table>    
  </div>
</template>