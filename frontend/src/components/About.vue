<script>
export default {
  data() {
    return {
    }
  },
  props: ['versioninfo', 'clienttz'],
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
            <td>API-Version</td>
            <td>{{ versioninfo }}</td>
        </tr>
        <tr>
            <td>Zeitzone des Clients</td>
            <td>{{ clienttz }}</td>
        </tr>
    </table>    
  </div>
</template>