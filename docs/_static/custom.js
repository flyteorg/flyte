// Add event listener for DOMContentLoaded event
window.addEventListener("DOMContentLoaded", function() {
    // Select all <a> elements with class "external"
    var externalLinks = document.querySelectorAll("a.external");

    // Loop through each <a> element with class "external"
    externalLinks.forEach(function(link) {
        // Set the target attribute to "_blank"
        link.setAttribute("target", "_blank");
    });
});

// This function was adapted from pydata-sphinx-theme
// https://github.com/pydata/pydata-sphinx-theme/blob/733d9f3264020c8a5bd3dde38f3ee3e5cdb2979a/src/pydata_sphinx_theme/assets/scripts/pydata-sphinx-theme.js#L133-L175
function scrollToActive() {
    // If the docs nav doesn't exist, do nothing (e.g., on search page)
    if (!document.querySelector(".sidebar-scroll")) {
      return;
    }

    var sidebar = document.querySelector("div.sidebar-scroll");

    // Remember the sidebar scroll position between page loads
    // Inspired on source of revealjs.com
    let storedScrollTop = parseInt(
      sessionStorage.getItem("sidebar-scroll-top"),
      10
    );

    if (!isNaN(storedScrollTop)) {
      // If we've got a saved scroll position, just use that
      sidebar.scrollTop = storedScrollTop;
      console.log("Scrolled sidebar using stored browser position...");
    } else {
      // Otherwise, calculate a position to scroll to based on the lowest `active` link
      var sidebarNav = document.querySelector(".sidebar-scroll");
      var active_pages = sidebarNav.querySelectorAll(".current-page");
      if (active_pages.length > 0) {
        // Use the last active page as the offset since it's the page we're on
        var latest_active = active_pages[active_pages.length - 1];
        var offset =
          latest_active.getBoundingClientRect().y -
          sidebar.getBoundingClientRect().y;
        // Only scroll the navbar if the active link is lower than 50% of the page
        if (latest_active.getBoundingClientRect().y > window.innerHeight * 0.5) {
          let buffer = 0.25; // Buffer so we have some space above the scrolled item
          sidebar.scrollTop = offset - sidebar.clientHeight * buffer;
          console.log("Scrolled sidebar using last active link...");
        }
      }
    }

    setTimeout(function() {
      // sidebar is hidden by default, so we need to make it visible
      // after scrolling. This prevents the scrollbar from jittering when
      // the page loads.
      console.log("Sidebar is now visible...")
      sidebar.style.visibility = "visible";
    }, 10);

    // Store the sidebar scroll position
    window.addEventListener("beforeunload", () => {
      sessionStorage.setItem("sidebar-scroll-top", sidebar.scrollTop);
    });
  }

document.addEventListener('DOMContentLoaded', scrollToActive);
