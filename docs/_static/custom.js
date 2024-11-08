window.addEventListener("DOMContentLoaded", function() {
    // Select all <a> elements with class "external"
    var externalLinks = document.querySelectorAll("a.external");

    // Loop through each <a> element with class "external"
    externalLinks.forEach(function(link) {
        // Set the target attribute to "_blank"
        link.setAttribute("target", "_blank");
    });
});
