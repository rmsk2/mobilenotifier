<script>

import PopupModal from './PopupModal.vue'
import { DaySelectorResult, sundayLast, daysPerMonth } from './globals';

const monthNames = ["Januar", "Februar", "März", "April", "Mai", "Juni", "Juli", "August", "September", "Oktober", "November", "Dezember"]

export default {
  components: { PopupModal },
  data() {
    return {
      title: "",
      cancelButton: "Abbrechen",
      selectedDay: 1,
      weekDayFirstInMonth: 0,
      allDays: [],

      resolvePromise: undefined,
      rejectPromise: undefined,
    }
  },
  methods: {
    show(day, month, year) {
      this.title = `Bitte Tag im ${monthNames[month-1]} ${year} auswählen`;
      this.selectedDay = day;
      this.month = month;
      this.year = year;
      this.monthNames = monthNames

      this.weekDayFirstInMonth = sundayLast(new Date(year, month-1, 1).getDay());

      this.allDays = []
      for (let row = 0; row < 6; row++) {
        for (let column = 0; column < 7; column++) {
          this.allDays.push("")
        }
      }

      let dayCount = 1;
      let offset = this.weekDayFirstInMonth;
      let daysInMonth = daysPerMonth(month, year)
      
      while (dayCount <= daysInMonth) {
        this.allDays[offset + (dayCount - 1)] = `${dayCount}`;
        dayCount++;
      }

      this.$refs.popup.open();
      
      return new Promise((resolve, reject) => {
        this.resolvePromise = resolve
        this.rejectPromise = reject
      })
    },
    weekN(weekNum) {
      return this.allDays.slice(weekNum * 7, (weekNum + 1) * 7);
    },
    changeSelected(dayNum) {
      this.selectedDay = dayNum;
      this.confirm()
    },
    confirm() {
      this.$refs.popup.close();
      this.resolvePromise(new DaySelectorResult(true, this.selectedDay));
    },
    cancel() {
      this.$refs.popup.close();
      this.resolvePromise(new DaySelectorResult(false, 0));
    },
  },
}
</script>

<template>
  <popup-modal ref="popup">
    <h3 style="margin-top: 0" class="conf-header">{{ title }}</h3>
      <div class="center-dayselect"><p>Aktuell ausgewählt: {{ selectedDay }} ter {{ monthNames[this.month - 1]}} {{ this.year }}</p></div>
      <div class="center-dayselect"><table class="daySelectorTable">
        <tr>
          <th class="dayselect-cell-right">Mo</th>
          <th class="dayselect-cell-right">Di</th>
          <th class="dayselect-cell-right">Mi</th>
          <th class="dayselect-cell-right">Do</th>
          <th class="dayselect-cell-right">Fr</th>
          <th class="dayselect-cell-right">Sa</th>
          <th class="dayselect-cell-right">So</th>
        </tr>
        <tr v-for="j in [0, 1, 2, 3, 4, 5]">
          <td class="dayselect-cell-right" v-for="i in weekN(j)">
            <div v-if="i===''"></div>
            <button class="dateselectbutton" v-else @click="changeSelected(i)">{{ i }}</button>
          </td>
        </tr>
      </table></div>
      <p/>
      <div class="center-dayselect"><button @click="cancel">{{ cancelButton }}</button></div>
  </popup-modal>
</template>

<style scoped>
.center-dayselect {
  display: flex;
  justify-content: center;
  align-items: center;
}
.dayselect-cell-right{
  text-align: right;
}
</style>