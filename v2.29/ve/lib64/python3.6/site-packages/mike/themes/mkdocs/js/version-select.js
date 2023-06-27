window.addEventListener("DOMContentLoaded", function() {
  function expandPath(path) {
    // Get the base directory components.
    var expanded = window.location.pathname.split("/");
    expanded.pop();
    var isSubdir = false;

    path.split("/").forEach(function(bit, i) {
      if (bit === "" && i === 0) {
        isSubdir = false;
        expanded = [""];
      } else if (bit === "." || bit === "") {
        isSubdir = true;
      } else if (bit === "..") {
        if (expanded.length === 1) {
          // We must be trying to .. past the root!
          throw new Error("invalid path");
        } else {
          isSubdir = true;
          expanded.pop();
        }
      } else {
        isSubdir = false;
        expanded.push(bit);
      }
    });

    if (isSubdir)
      expanded.push("");
    return expanded.join("/");
  }

  // `base_url` comes from the base.html template for this theme.
  var ABS_BASE_URL = expandPath(base_url);
  var CURRENT_VERSION = ABS_BASE_URL.match(/\/([^\/]+)\/$/)[1];

  function makeSelect(options, selected) {
    var select = document.createElement("select");
    select.classList.add("form-control");

    options.forEach(function(i) {
      var option = new Option(i.text, i.value, undefined,
                              i.value === selected);
      select.add(option);
    });

    return select;
  }

  var xhr = new XMLHttpRequest();
  xhr.open("GET", ABS_BASE_URL + "../versions.json");
  xhr.onload = function() {
    var versions = JSON.parse(this.responseText);

    var realVersion = versions.find(function(i) {
      return i.version === CURRENT_VERSION ||
             i.aliases.includes(CURRENT_VERSION);
    }).version;

    var select = makeSelect(versions.map(function(i) {
      return {text: i.title, value: i.version};
    }), realVersion);
    select.addEventListener("change", function(event) {
      window.location.href = ABS_BASE_URL + "../" + this.value + "/";
    });

    var container = document.createElement("div");
    container.id = "version-selector";
    container.appendChild(select);

    var title = document.querySelector(".navbar-brand");
    if (title.parentNode.classList.contains("navbar-header")) {
      var height = window.getComputedStyle(title).getPropertyValue("height");
      container.style.height = height;
    }

    title.parentNode.insertBefore(container, title.nextSibling);
  };
  xhr.send();
});
