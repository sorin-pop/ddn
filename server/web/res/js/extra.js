$(document).ready(function() {
    checkConnector()
})

$(document).on('change', '#connector', function() {
    checkConnector()
});
    

function checkConnector() {
    var connector = $("#connector")
    
    if (connector.length != 0) {
        if (connector.val().includes("oracle")) {
            $("#dbname").prop('disabled', true);
        } else {
            $("#dbname").prop('disabled', false);
        }
    }
}