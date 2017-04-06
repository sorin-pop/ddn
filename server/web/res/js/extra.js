$(document).ready(function() {
    // Enable tooltips
    $(function () {
        $('[data-toggle="tooltip"]').tooltip()
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
        if (connector.val().toLowerCase().includes("oracle")) {
            $("#dbname").prop('disabled', true);
            $("#dbnamediv").attr('title', 'Not needed for Oracle. Think of the User field below as the "database", as it will also be the Oracle schema that will contain the tables and their data.').tooltip('show');
            $("#userdiv").attr('title','');
            $("#userdiv").tooltip('hide');
        } else {
            $("#dbname").prop('disabled', false);
            $("#dbnamediv").attr('title', '');
        }

        if (connector.val().toLowerCase().includes("mssql") || connector.val().toLowerCase().includes("sql server")) {
            $("#user").prop('disabled', true);
            $("#password").prop('disabled', true);
            $("#userdiv").attr('title', 'User and password not needed for SQL Server.').tooltip('show');
            $("#dbnamediv").attr('title','');
            $("#dbnamediv").tooltip('hide');
        } else {
            $("#user").prop('disabled', false);
            $("#password").prop('disabled', false);
            $("#userdiv").attr('title', '');
        }
    }
}