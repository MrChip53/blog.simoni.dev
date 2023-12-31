(function() {
    htmx.defineExtension("theme", {
        onEvent: function (name, evt) {
            if (name === "htmx:afterSettle") {
                const theme = evt.detail.xhr.getResponseHeader("HX-Theme");
                if (!!theme) {
                    document.getElementById("theme").setAttribute("href", `/css/themes/${theme}.css`);
                }
            }
        }
    });
})();