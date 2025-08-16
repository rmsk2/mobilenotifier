const reminderAnniversary = 1
const reminderOneShot = 2
const reminderMonthly = 3

const warningMorningBefore = 1
const warningNoonBefore = 2
const warningEveningBefore = 3
const warningWeekBefore = 4
const warningSameDay = 5

const addrClassIfttt = "IFTTT"
const addrClassMail = "Mail"

export { 
    ReminderAPI, APIResult, Reminder, ReminderData, ReminderResponse, ReminderOverview,
    ExtReminder, ReminderListResponse, OverviewResponse, ApiInfoResult, RecipientInfo, getDefaultReminder,
    RecipientData, Recipient,
    reminderAnniversary, reminderOneShot, reminderMonthly,
    warningMorningBefore, warningNoonBefore, warningEveningBefore, warningWeekBefore, warningSameDay,
    addrClassIfttt, addrClassMail
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

class ApiInfoResult {
    constructor(version, timeZone, elemCount, metrics) {
        this.version_info = version;
        this.time_zone = timeZone;
        this.reminder_count = elemCount;
        this.metrics = metrics
    }
}

class RecipientInfo {
    constructor(displayName, id) {
        this.display_name = displayName;
        this.id = id;
    }
}

class RecipientData {
    constructor(adrType, adr, display_name, is_default) {
        this.addr_type = adrType
        this.address = adr
        this.display_name = display_name
        this.is_default = is_default
    }
}

class Recipient extends RecipientData {
    constructor(adrType, adr, display_name, id, is_default) {
        super(adrType, adr, display_name, is_default)
        this.id = id
    }

    toData() {
        return new RecipientData(this.addr_type, this.address, this.display_name, this.is_default)
    }
}

function getDefaultReminder(recipients) {
    let now = new Date();
    return new Reminder(null, reminderOneShot, 0, [warningSameDay], now, "Neues Ereignis", recipients);
}

function timeout(ms) {
    return new Promise(resolve => setTimeout(resolve, ms));
}

class ReminderAPI {
    constructor(baseUrl, accessToken) {
        this.URL = baseUrl
        this.Token = accessToken
    }

    getURL() {
        return this.URL;
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

    async upsertAddressBookEntry(entryData, apiUrl, method) {
        try
        {
            let response = await fetch(apiUrl, {
                method: method,
                headers: {
                    'Content-Type': 'application/json',
                    'Accept': 'application/json',
                },
                body: JSON.stringify(entryData)
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

    async updateAddressBookEntry(entryData, id) {
        let apiUrl = `${this.URL}addressbook/${id}`;

        return await this.upsertAddressBookEntry(entryData, apiUrl, "put");
    }

    async newAddressBookEntry(entryData) {
        let apiUrl = `${this.URL}addressbook`;

        return await this.upsertAddressBookEntry(entryData, apiUrl, "post");
    }

    async deleteById(id, baseUrl) {
        try
        {
            let apiUrl = `${baseUrl}${id}`;
            
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

    async deleteReminder(id) {
        let baseUrl = `${this.URL}reminder/`;

        return await this.deleteById(id, baseUrl);
    }

    async deleteAddressBookEntry(id) {
        let baseUrl = `${this.URL}addressbook/`;

        return await this.deleteById(id, baseUrl);
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

            return new APIResult(false, allRecipients)
        } catch(error) {
            return new APIResult(true, error);
        }        
    }

    async getFullRecipients() {
        try
        {
            let apiUrl = `${this.URL}addressbook`;

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

            return new APIResult(false, allRecipients)
        } catch(error) {
            return new APIResult(true, error);
        }
    }

    async getApiInfo() {
        try
        {
            let apiUrl = `${this.URL}general/info`;
            
            let response = await fetch(apiUrl, {
                method: "get",
                headers: {
                    'Accept': 'application/json',
                }
            });

            if (!response.ok) {
                return new APIResult(true, `${response.status}`);
            }

            let apiInfo = await response.json();            

            return new APIResult(false, apiInfo)
        } catch(error) {
            return new APIResult(true, error);
        }        
    }    
}