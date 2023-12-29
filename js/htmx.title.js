(function() {
    htmx.defineExtension("title", {
        onEvent: function (name, evt) {
            if (name === "htmx:afterOnLoad") {
                const titleHeader = evt.detail.xhr.getResponseHeader("HX-Title");
                if (!!titleHeader) {
                    document.title = titleHeader;
                }
            }
        }
    });
})();