const monthSelected = 1;
const newSelected = 2;
const allSelected = 3;
const aboutSelected = 4;
const recipientListSelected = 5;
const versionString = "1.3.3";

class DeleteNotification {
   constructor(id, description) {
    this.id = id;
    this.description = description;
  }
}

class DaySelectorResult {
  constructor(success, day) {
    this.success = success;
    this.day = day;
  }
}

export { 
  monthSelected, newSelected, allSelected, aboutSelected, versionString, recipientListSelected, DaySelectorResult,
  DeleteNotification, isLeapYear, incDay, decDay, sucMonth, predMonth, performDateCorrection, sucYear, predYear,
  daysPerMonth, sundayLast
};


function isLeapYear(year) {
  return ((year % 4 == 0) && (year % 100 != 0)) || (year % 400 == 0);
}

function sundayLast(dayNum) {
  if (dayNum === 0) {
    return 6
  } else {
    return dayNum - 1
  }
}

var m30 = new Set([3, 5, 8, 10])

function incDay(day, month, year) {
  let m = month - 1
  let y = year
  let d = day + 1

  if (m === 1) {
    // February
    let maxDay = 28

    if (isLeapYear(y)) {
      maxDay = 29
    }

    if (d > maxDay) {
      m++
      d = 1
    }
  } else if (m === 11) {
    // December
    if (d > 31) {
      m = 0
      d = 1
      y++
    }
  } else {
    let maxDay = 31
    if (m30.has(m)) {
      maxDay = 30
    }

    if (d > maxDay) {
      m = m + 1
      d = 1
    }
  }

  return {day: d, month: m+1, year: y}
}

function decDay(day, month, year) {
  let m = month - 1
  let y = year
  let d = day - 1

  if (d === 0) {
    if (m === 2) {
      // March to February
      let febDay = 28
      if (isLeapYear(year)) {
        febDay = 29
      }

      m = 1
      d = febDay
    } else if (m === 0) {
      // January to December
      d = 31
      m = 11
      y--
    } else {
      // All other months
      m--
      d = 31

      if (m30.has(m)) {
        d = 30
      }
    }
  }

  return {day: d, month: m+1, year: y}
}

function sucMonth(day, month, year) {
  let d = day
  let y = year
  let m = month - 1

  m = (m + 1) % 12
  if (m === 0) {
    y++
  }

  m++
  return {day: performDateCorrection(d, m, y), month: m, year: y}
}

function sucYear(day, month, year) {
  let y = year + 1
  let d = day
  let m = month

  return {day: performDateCorrection(d, m, y), month: m, year: y}
}

function predYear(day, month, year) {
  let y = year - 1
  if (y < 0) {
    y = 0
  }
  let d = day
  let m = month

  return {day: performDateCorrection(d, m, y), month: m, year: y}  
}


function predMonth(day, month, year) {
  let d = day
  let y = year
  let m = month - 1

  m = (m + 11) % 12
  if (m === 11) {
    y--
  }

  m++
  return {day: performDateCorrection(d, m, y), month: m, year: y}
}

function daysPerMonth(month, year) {
  let maxDay = 31;

  if (month == 2) {
    maxDay = 28
    if (isLeapYear(year)) {
      maxDay = 29;
    }
  } else {
    if (m30.has(month - 1)) {
      maxDay = 30;
    }
  }

  return maxDay;
}

function performDateCorrection(day, month, year) {
  let maxDay = daysPerMonth(month, year)

  if (day > maxDay) {
    day = maxDay
  }

  return day
}