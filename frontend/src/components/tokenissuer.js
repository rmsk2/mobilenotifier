export { IssuerAPI };

class IssuerAPI {
    constructor(issuerUrl) {
        this.URL = issuerUrl
    }

    async getToken() {
        let reminderData = {"audience": "gschmarri"};
        
        try
        {            
            let response = await fetch(this.URL, {
                method: "post",
                headers: {
                    'Content-Type': 'application/json',
                    'Accept': 'application/json',
                },
                body: JSON.stringify(reminderData)
            });

            if (!response.ok) {
                return "notoken";
            }

            let result = await response.json();

            return result.token;
        } catch(error) {
            return "notoken";
        }        
    }
}