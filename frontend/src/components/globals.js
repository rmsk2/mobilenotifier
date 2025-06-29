const monthSelected = 1;
const newSelected = 2;
const allSelected = 3;
const aboutSelected = 4;
const versionString = "0.5.17";

class DeleteNotification {
    constructor(id, description) {
        this.id = id;
        this.description = description;
    }
}

export { 
    monthSelected, newSelected, allSelected, aboutSelected, versionString,
    DeleteNotification
 };
