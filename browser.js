"use strict";
var page = require('webpage').create();
var system = require('system');

page.onConsoleMessage = function(msg) {
    console.log(msg);
};

page.open(system.args[1], function(status) {
    if (status === "success") {
        page.evaluate(function() {
            // lastest version on the web
            console.log(document.documentElement.outerHTML);
        });
        phantom.exit(0);
    } else {
      phantom.exit(1);
    }
});
