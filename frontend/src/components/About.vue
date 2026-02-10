<script>
import { versionString } from './globals';

const showExtendedInfoText = "Mehr Info";
const showSimpleInfoText = "Weniger Info";

export default {
  data() {
    return {
      webUIVersion: versionString,
      reminderCount: 0,
      notificationCount: 0,
      showExtendedInfo: false,
      infoButtonText: showSimpleInfoText
    }
  },
  props: ['versioninfo', 'clienttz', 'apilink', 'elemcount', 'metrics'],
  emits: ['copy-token'],
  methods: {
    copyTokenClicked() {
      this.$emit('copy-token')
    },
    toggleInfo() {
      this.showExtendedInfo = !this.showExtendedInfo;
    }
  },
  computed: {
    swaggerUrl() {
      return `${this.apilink}swagger/index.html`;
    },
    buttonText() {
      if (this.showExtendedInfo) {
        return showSimpleInfoText;
      } else {
        return showExtendedInfoText;
      }
    }
  },
  watch: {
    elemcount: {
        handler(newVal){
          this.reminderCount = newVal;
        },
        immediate: true
    },
    metrics: {
        handler(newVal){
          this.notificationCount = newVal;
        },
        immediate: true
    }    
  },
}
</script>

<template>
  <div id="div-about" class="about-component">
    <h2>Über diese Anwendung</h2>
    <div class="about-text">
    Mobile-Notifier (a.k.a das Gschmarri-Projekt). Eine Webanwendung zur Verwaltung von SMS-Erinnerungen. Geschrieben von Martin Grap in 2025 und 2026.
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
      <tr>
        <td class="table-about-elem"><span class="about-text">Anzahl Ereignisse</span></td>
        <td class="table-about-elem"><span class="about-text">{{ reminderCount }}</span></td>
      </tr>
      <tr>
        <td class="table-about-elem"><span class="about-text">Anzahl Benachrichtigungen</span></td>
        <td class="table-about-elem"><span class="about-text">{{ metrics.notification_count }}</span></td>
      </tr>      
    </table>
    <p/>
    <button @click="toggleInfo()">{{ buttonText }}</button>
    <p v-if="showExtendedInfo">
      <table class="about-table">
          <tr>
          <th class="table-about-header"><span class="about-text">Info</span></th>
          <th class="table-about-header"><span class="about-text">Wert</span></th>
        </tr>
        <tr v-for="(v, k) in metrics">
            <td class="table-about-elem"><span class="about-text">{{ k }}</span></td>
            <td class="table-about-elem"><span class="about-text">{{ v }}</span></td>
        </tr>
      </table>
    </p>
    <p/>
    Wenn ein Token für das Backupscript oder für die direkte Bedienung des APIs via Swagger benötigt wird, kann der aktuell verwendete Token über
    den unten stehenden Button in die Zwischenablage kopiert werden.
    <p/>
    <button @click="copyTokenClicked()">Aktuellen Token kopieren</button>
  </div>
</template>