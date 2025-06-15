const reminderAnniversary = 1
const reminderOneShot = 2

const warningMorningBefore = 1
const warningNoonBefore = 2
const warningEveningBefore = 3
const warningWeekBefore = 4
const warningSameDay = 5

export { 
    ReminderAPI, APIResult, Reminder, ReminderData, ReminderResponse, ReminderOverview,
    ExtReminder, ReminderListResponse, OverviewResponse, getDefaultReminder,
    reminderAnniversary, reminderOneShot,
    warningMorningBefore, warningNoonBefore, warningEveningBefore, warningWeekBefore, warningSameDay
 };

class ReminderResponse {
    constructor(wasFound, reminder) {
        this.found = wasFound;
        this.data = reminder;
    }
}

class SmallReminder {
    constructor(id, description, kind) {
        this.id = id;
        this.description = description;
        this.kind = kind;
    }    
}

class ReminderOverview {
    constructor(id, description, kind, nextEvent) {
        this.reminder = new SmallReminder(id, description, kind)
        this.next_occurrance = nextEvent
    }    
}

class OverviewResponse {
    constructor(reminderOverviews) {
        this.reminders = reminderOverviews
    }
}

class ReminderData {
    constructor(kind, param, warningAt, spec, description, recipients) {
        this.kind = kind;
        this.param = param;
        this.warning_at = warningAt;        
        this.spec = spec;
        this.description = description;
        this.recipients = recipients;
    }
}

class Reminder extends ReminderData {
    constructor(id, kind, param, warningAt, spec, description, recipients) {
        super(kind, param, warningAt, spec, description, recipients)
        this.id = id;
    }
}

class ExtReminder {
    constructor(reminder, nextEvent) {
        this.reminder = reminder;
        this.next_occurrance = nextEvent;
    }
}

class ReminderListResponse {
    constructor(extReminders) {
        this.reminders = extReminders
    }
}


class APIResult {
    constructor(wasError, data) {
        this.error = wasError;
        this.data = data;
    }    
}

function getDefaultReminder(recipient) {
    let now = new Date();
    return new Reminder(null, reminderOneShot, 0, [warningSameDay], now, "Neues Ereignis", [recipient]);
}

class ReminderAPI {
    constructor(baseUrl, accessToken) {
        this.URL = baseUrl
        this.Token = accessToken
    }

    async createNewReminder(reminderData) {
        try
        {
            let apiUrl = `${this.URL}reminder`;
            
            let response = await fetch(apiUrl, {
                method: "post",
                headers: {
                    'Content-Type': 'application/json',
                    'Accept': 'application/json',
                },
                body: JSON.stringify(reminderData)
            });

            if (!response.ok) {
                return new APIResult(true, `${response.status}`);
            }

            let result = await response.json();

            return new APIResult(false, result.uuid);
        } catch(error) {
            return new APIResult(true, error);
        }        
    }

    async readReminder(id) {
        try
        {
            let apiUrl = `${this.URL}reminder/${id}`;
            
            let response = await fetch(apiUrl, {
                method: "get",
                headers: {
                    'Accept': 'application/json',
                }
            });

            if (!response.ok) {
                return new APIResult(true, `${response.status}`);
            }

            let result = await response.json();

            return new APIResult(false, result);
        } catch(error) {
            return new APIResult(true, error);
        }    
    }

    async updateReminder(reminderData, id) {
        try
        {
            let apiUrl = `${this.URL}reminder/${id}`;
            
            let response = await fetch(apiUrl, {
                method: "put",
                headers: {
                    'Content-Type': 'application/json',
                    'Accept': 'application/json',
                },
                body: JSON.stringify(reminderData)
            });

            if (!response.ok) {
                return new APIResult(true, `${response.status}`);
            }

            let result = await response.json();

            return new APIResult(false, result.uuid);
        } catch(error) {
            return new APIResult(true, error);
        }
    }

    async deleteReminder(id) {
        try
        {
            let apiUrl = `${this.URL}reminder/${id}`;
            
            let response = await fetch(apiUrl, {
                method: "delete",
                headers: {
                    'Accept': 'application/json',
                }
            });

            if (response.ok) {
                return new APIResult(false, "")
            } else {
                return new APIResult(true, `${response.status}`);
            }
        } catch(error) {
            return new APIResult(true, error);
        }
    }

    async getOverview() {
        try
        {
            let apiUrl = `${this.URL}reminder/views/basic`;
            
            let response = await fetch(apiUrl + "?" + new URLSearchParams({max_entries: 0}).toString(), {
                method: "get",
                headers: {
                    'Accept': 'application/json',
                }
            });

            if (!response.ok) {
                return new APIResult(true, `${response.status}`);
            }

            let overview = await response.json();            

            return new APIResult(false, overview.reminders)
        } catch(error) {
            return new APIResult(true, error);
        }        
    }

    async getEventsInMonth(m, y) {
        try
        {
            let apiUrl = `${this.URL}reminder/views/bymonth`;
            
            let response = await fetch(apiUrl + "?" + new URLSearchParams({year: y, month: m}).toString(), {
                method: "get",
                headers: {
                    'Accept': 'application/json',
                }
            });

            if (!response.ok) {
                return new APIResult(true, `${response.status}`);
            }

            let overview = await response.json();            

            return new APIResult(false, overview.reminders)
        } catch(error) {
            return new APIResult(true, error);
        }        
    }

    async sendSms(messageTxt, recipient) {
        try
        {
            let apiUrl = `${this.URL}send/${recipient}`;
            
            let response = await fetch(apiUrl, {
                method: "post",
                headers: {
                    'Accept': 'application/json',
                    'Content-Type': 'application/json',
                    'X-Token': this.Token
                },
                body: JSON.stringify({message: messageTxt})
            });

            if (response.ok) {
                return new APIResult(false, "")
            } else {
                return new APIResult(true, `${response.status}`);
            }
        } catch(error) {
            return new APIResult(true, error);
        }
    }

    async getRecipients() {
        try
        {
            let apiUrl = `${this.URL}send/recipients/all`;
            
            let response = await fetch(apiUrl, {
                method: "get",
                headers: {
                    'Accept': 'application/json',
                }
            });

            if (!response.ok) {
                return new APIResult(true, `${response.status}`);
            }

            let allRecipients = await response.json();            

            return new APIResult(false, allRecipients.all_recipients)
        } catch(error) {
            return new APIResult(true, error);
        }        
    }
}