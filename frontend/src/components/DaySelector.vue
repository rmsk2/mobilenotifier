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
        this.allDays[offset + (dayCount - 1)] = dayCount;
        dayCount++;
      }

      this.$refs.popup.open();
      
      return new Promise((resolve, reject) => {
        this.resolvePromise = resolve
        this.rejectPromise = reject
      })
    },
    calcClass(day) {
      if (day === this.selectedDay) {
        return "dateselectbutton-current"
      } else {
        return "dateselectbutton"
      }
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
          <th class="dayselect-cell">Mo</th>
          <th class="dayselect-cell">Di</th>
          <th class="dayselect-cell">Mi</th>
          <th class="dayselect-cell">Do</th>
          <th class="dayselect-cell">Fr</th>
          <th class="dayselect-cell">Sa</th>
          <th class="dayselect-cell">So</th>
        </tr>
        <tr v-for="j in [0, 1, 2, 3, 4, 5]">
          <td class="dayselect-cell" v-for="i in weekN(j)">
            <div v-if="i===''"></div>
            <button :class="calcClass(i)" v-else @click="changeSelected(i)">{{ i }}</button>
          </td>
        </tr>
      </table></div>
      <p/>
      <div class="center-dayselect"><button @click="cancel">{{ cancelButton }}</button></div>
  </popup-modal>
</template>
