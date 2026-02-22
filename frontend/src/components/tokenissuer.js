export { IssuerAPI };

class IssuerAPI {
    constructor(issuerUrl, audience, algorithm) {
        this.URL = issuerUrl
        this.audience = audience
        this.algorithm = algorithm
    }

    async getToken() {
        let reminderData = {"audience": this.audience, "algorithm": this.algorithm};
        
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