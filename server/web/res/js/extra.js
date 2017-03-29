$(document).ready(function() {
    checkConnector()
    
    $("#dbname").keyup(function() {
        checkInputs()
    });

    $("#user").keyup(function() {
        checkInputs()
    })

})

$(document).on('change', '#connector', function() {
    checkConnector()
});

function checkInputs() {
    if (valid("#dbname") && valid("#user")) {
        $("button").prop("disabled", false)
    } else {
        $("button").prop("disabled", true)
    }
}

function valid(selector) {
    var value = $(selector).val()
    var pattern = "^[a-zA-Z0-9$_]+$"

    if (value.match(pattern) || value == "") {
        $(selector).parent().removeClass("has-danger")
        $(selector).removeClass("form-control-danger")

        return true
    }

    $(selector).parent().addClass("has-danger")
    $(selector).addClass("form-control-danger")

    return false
}

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