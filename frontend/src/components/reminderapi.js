export { ReminderAPI };

class APIResult {
    constructor(wasError, data) {
        this.error = wasError;
        this.data = data;
    }    
}

class ReminderAPI {
    constructor(baseUrl, accessToken) {
        this.URL = baseUrl
        this.Token = accessToken
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