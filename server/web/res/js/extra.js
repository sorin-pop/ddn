$(document).ready(function() {
    // Enable tooltips
    $(function() {
        $('[data-toggle="tooltip"]').tooltip()
        $('[data-toggle="popover"]').popover()
    })


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
        $("#dbname").prop('disabled', false);
        $("#user").prop('disabled', false);
        $("#password").prop('disabled', false);
        $("#dbnamediv").attr('title', '').attr('data-original-title', '').tooltip('hide');
        $("#userdiv").attr('title', '').attr('data-original-title', '').tooltip('hide');

        connectorVal = connector.val().toLowerCase();

        if (connectorVal.includes("oracle")) {
            msg = 'Not needed for Oracle. Think of the User field below as the "database", as it will also be the Oracle schema that will contain the tables and their data.';

            $("#dbname").prop('disabled', true);
            $("#dbnamediv").attr('title', msg).attr('data-original-title', msg).tooltip('hide');
        } else if (connectorVal.includes("mssql") || connectorVal.includes("sql server")) {
            msg = 'User and password not needed for SQL Server.'

            $("#user").prop('disabled', true);
            $("#password").prop('disabled', true);

            $("#userdiv").attr('title', msg).attr('data-original-title', msg).tooltip('hide');
        }
    }
}